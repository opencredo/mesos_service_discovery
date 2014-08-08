package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

type Event struct {
  EventType string
  TaskStatus string
  AppId string
  TaskId string
  Host string
  Ports []int
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
  Id string
  Ports []int
  ApplicationInstances map[string]Task
}

type Task struct {
  Id string
  Host string
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
var applicationMap = make(map[string]Application)

var MARATHON_HOST = "172.16.5.10"
var MARATHON_PORT = "8080"

func loadExistingTasks(appId string) {
  resp, err := http.Get("http://" + MARATHON_HOST + ":" + MARATHON_PORT + "/v2/apps/" + appId)
  if (err != nil) { return; }
  body, err2 := ioutil.ReadAll(resp.Body)
  if (err2 != nil) { return; }
  var app MarathonTasksResponse
  err3 := json.Unmarshal(body, &app)
  if (err3 != nil) { return; }

  for _, task := range app.App.Tasks {
    addTask(appId, task.Host, task.Ports, task.Id)
  }
}

func loadExistingApps() {
  resp, err := http.Get("http://" + MARATHON_HOST + ":" + MARATHON_PORT + "/v2/apps")
  if (err != nil) { return; }
  body, err2 := ioutil.ReadAll(resp.Body)
  if (err2 != nil) { return; }
  var applications MarathonAppsResponse
  err3 := json.Unmarshal(body, &applications)
  if (err3 != nil) { return; }

  for _, app := range applications.Apps {
    _, ok := applicationMap[app.Id]
    if (!ok) {
      fmt.Printf("Found application: %s\n", app.Id)
      app.ApplicationInstances = make(map[string]Task)
      applicationMap[app.Id] = app
      loadExistingTasks(app.Id)
    }
  }
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if (err == nil) { eventQueue <- body }
}

func parseEvent(event []byte) (Event, bool) {
  var e Event
  err := json.Unmarshal(event, &e)
  return e, err == nil
}

func addTask(appId string, host string, ports []int, taskId string) {
  task := Task{taskId, host, ports}
  app, ok := applicationMap[appId]
  if (!ok) { loadExistingApps() }
  app, ok = applicationMap[appId]
  if (!ok) {
    fmt.Printf("ERROR Unknown application %s\n", appId)
    return
  }
  app.ApplicationInstances[taskId] = task
  fmt.Printf("Found task for %s on %s [%s]\n", appId, task.Host, task.Id)
}

func removeTask(appId string, host string, ports []int, taskId string) {
  app := applicationMap[appId]
  delete(app.ApplicationInstances, taskId)
  fmt.Printf("Removed task for %s on %s [%s]\n", appId, host, taskId) 
}

func generateHAProxyConfig() {
  tmp, err := ioutil.TempFile("", "haproxy.cfg")
  if (err != nil) { return; }

  fmt.Fprintf(tmp, haproxyHeader)
  for appId, app := range applicationMap {
    fmt.Fprintf(tmp, "\nlisten %s\n  bind 0.0.0.0:%d\n  mode tcp\n  option tcplog\n  balance leastconn\n", appId, app.Ports[0])
    i := 0
    for _, task := range app.ApplicationInstances {
      fmt.Fprintf(tmp, "  server %s-%d %s:%d check\n", appId, i, task.Host, task.Ports[0])
      i++
    }
  }
  fmt.Println(tmp.Name())
}

func eventsWorker() {
  for {
    event := <-eventQueue
    e, ok := parseEvent(event)
    if (ok && e.EventType == "status_update_event") {
      if (e.TaskStatus == "TASK_RUNNING") {
        addTask(e.AppId, e.Host, e.Ports, e.TaskId)
      } else if (e.TaskStatus == "TASK_KILLED" || e.TaskStatus == "TASK_LOST" || e.TaskStatus == "TASK_FAILED") {
        removeTask(e.AppId, e.Host, e.Ports, e.TaskId)
      } else {
        fmt.Printf("WARN: Unknown task status %s\n", e.TaskStatus);
      }
      generateHAProxyConfig();
    }
  }
}

func main() {
  fmt.Printf("Running things and stuff\n")
  loadExistingApps()
  generateHAProxyConfig()
  go eventsWorker()
  http.HandleFunc("/events", eventsHandler)
  http.ListenAndServe(":8080", nil)
}
