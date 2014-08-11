package main

import (
	"encoding/json"
	"flag"
	"fmt"
  "io/ioutil"
	"net/http"
)

type MarathonAppsResponse struct {
	Apps []Application
}

type MarathonApp struct {
	Tasks []Task
}

type MarathonTasksResponse struct {
	App MarathonApp
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

func getMarathonAddress() string {
  return "http://" + *MARATHON_HOST + ":" + *MARATHON_PORT
}

func getResponseJSON(address string, v interface{}) error {
	resp, err := http.Get(address)
	if err != nil {
		return err
	}
	body, err2 := ioutil.ReadAll(resp.Body)
  resp.Body.Close()
	if err2 != nil {
		return err2
	}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return err
	}
  return nil
}

func loadExistingTasks(applicationMap map[string]Application, appId string) {
  var app MarathonTasksResponse
  err := getResponseJSON(getMarathonAddress() + "/v2/apps/" + appId, &app)
	if err != nil {
		return
	}
	for _, task := range app.App.Tasks {
		addTask(applicationMap, appId, task)
	}
}

func loadExistingApps(applicationMap map[string]Application) {
	var applications MarathonAppsResponse
	err := getResponseJSON(getMarathonAddress() + "/v2/apps", &applications)
	if err != nil {
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
