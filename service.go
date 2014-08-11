package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

type Event struct {
	EventType  string
	TaskStatus string
	AppId      string
	TaskId     string
	Host       string
	Ports      []int
}

type MarathonAppsResponse struct {
	Apps []Application
}

type MarathonApp struct {
	Tasks []Task
}

type MarathonTasksResponse struct {
	App MarathonApp
}

type Application struct {
	Id                   string
	Ports                []int
	ApplicationInstances map[string]Task
}

type Task struct {
	Id    string
	Host  string
	Ports []int
}

var haproxyHeader = `
global
  daemon
  log 127.0.0.1 local0
  log 127.0.0.1 local1 notice
  maxconn 4096

defaults
  log         global
  retries     3
  maxconn     2000
  contimeout  5000
  clitimeout  50000
  srvtimeout  50000

listen stats
  bind 127.0.0.1:9090
  balance
  mode http
  stats enable
  stats auth admin:admin
`

var eventQueue = make(chan []byte)

var MARATHON_HOST = flag.String("host", "localhost", "The host Marathon is running on")
var MARATHON_PORT = flag.String("port", "8080", "The port Marathon is running on")

func loadExistingTasks(applicationMap map[string]Application, appId string) {
	resp, err := http.Get("http://" + *MARATHON_HOST + ":" + *MARATHON_PORT + "/v2/apps/" + appId)
	if err != nil {
		return
	}
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return
	}
	var app MarathonTasksResponse
	err3 := json.Unmarshal(body, &app)
	if err3 != nil {
		return
	}

	for _, task := range app.App.Tasks {
		addTask(applicationMap, appId, task.Host, task.Ports, task.Id)
	}
}

func loadExistingApps(applicationMap map[string]Application) {
	resp, err := http.Get("http://" + *MARATHON_HOST + ":" + *MARATHON_PORT + "/v2/apps")
	if err != nil {
		return
	}
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return
	}
	var applications MarathonAppsResponse
	err3 := json.Unmarshal(body, &applications)
	if err3 != nil {
		return
	}

	for _, app := range applications.Apps {
		_, ok := applicationMap[app.Id]
		if !ok {
			fmt.Printf("INFO Found application: %s\n", app.Id)
			app.ApplicationInstances = make(map[string]Task)
			applicationMap[app.Id] = app
			loadExistingTasks(applicationMap, app.Id)
		}
	}
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err == nil {
		eventQueue <- body
	}
}

func parseEvent(event []byte) (Event, bool) {
	var e Event
	err := json.Unmarshal(event, &e)
	return e, err == nil
}

func addTask(applicationMap map[string]Application, appId string, host string, ports []int, taskId string) {
	task := Task{taskId, host, ports}
	app, ok := applicationMap[appId]
	if !ok {
		loadExistingApps(applicationMap)
	}
	app, ok = applicationMap[appId]
	if !ok {
		fmt.Printf("ERR Unknown application %s\n", appId)
		return
	}
	app.ApplicationInstances[taskId] = task
	fmt.Printf("INFO Found task for %s on %s:%d [%s]\n", appId, task.Host, task.Ports[0], task.Id)
}

func removeTask(applicationMap map[string]Application, appId string, taskId string) {
	app := applicationMap[appId]
	delete(app.ApplicationInstances, taskId)
}

func generateHAProxyConfig(applicationMap map[string]Application) {
	tmp, err := ioutil.TempFile("", "haproxy.cfg")
	if err != nil {
		return
	}

	fmt.Fprintf(tmp, haproxyHeader)
	for appId, app := range applicationMap {
		fmt.Fprintf(tmp, "\nlisten %s\n  bind 0.0.0.0:%d\n  mode tcp\n  option tcplog\n  balance leastconn\n", appId, app.Ports[0])
		i := 0
		for _, task := range app.ApplicationInstances {
			fmt.Fprintf(tmp, "  server %s-%d %s:%d check\n", appId, i, task.Host, task.Ports[0])
			i++
		}
	}
	err = os.Rename(tmp.Name(), "/etc/haproxy/haproxy.cfg")
	if err != nil {
		fmt.Println("ERR Couldn't write /etc/haproxy/haproxy.cfg")
		fmt.Println(err)
		return
	}
	fmt.Println("INFO Written new /etc/haproxy/haproxy.cfg")
	cmd := exec.Command("service", "haproxy", "reload")
	err = cmd.Start()
	if err != nil {
		fmt.Println("ERR failed to reload HAProxy")
		return
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println("ERR failed to reload HAProxy")
		return
	}
}

func processStatusUpdateEvent(applicationMap map[string]Application, e Event) {
	if e.TaskStatus == "TASK_RUNNING" {
		addTask(applicationMap, e.AppId, e.Host, e.Ports, e.TaskId)
	} else if e.TaskStatus == "TASK_KILLED" || e.TaskStatus == "TASK_LOST" || e.TaskStatus == "TASK_FAILED" {
		removeTask(applicationMap, e.AppId, e.TaskId)
		fmt.Printf("INFO Removed task for %s on %s [%s]\n", e.AppId, e.Host, e.TaskId)
	} else {
		fmt.Printf("WARN Unknown task status %s\n", e.TaskStatus)
	}
}

func eventsWorker(applicationMap map[string]Application) {
	for {
		event := <-eventQueue
		e, ok := parseEvent(event)
		if ok && e.EventType == "status_update_event" {
			processStatusUpdateEvent(applicationMap, e)
			generateHAProxyConfig(applicationMap)
		}
	}
}

func main() {
	flag.Parse()
	fmt.Printf("Running things and stuff\n")
	applicationMap := make(map[string]Application)
	loadExistingApps(applicationMap)
	generateHAProxyConfig(applicationMap)
	go eventsWorker(applicationMap)
	http.HandleFunc("/events", eventsHandler)
	http.ListenAndServe(":8080", nil)
}
