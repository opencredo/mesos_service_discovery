package main

import (
  "encoding/json"
  "io/ioutil"
  "log"
  "net/http"
)

type Event struct {
  EventType  string
  TaskStatus string
  AppId      string
  TaskId     string
  Host       string
  Ports      []int
}

func processStatusUpdateEvent(applicationMap map[string]Application, e Event) {
  if e.TaskStatus == "TASK_RUNNING" {
    task := Task{e.TaskId, e.Host, e.Ports}
    addTask(applicationMap, e.AppId, task)
  } else if e.TaskStatus == "TASK_KILLED" || e.TaskStatus == "TASK_LOST" || e.TaskStatus == "TASK_FAILED" {
    removeTask(applicationMap, e.AppId, e.TaskId)
    log.Printf("INFO Removed task for %s on %s [%s]\n", e.AppId, e.Host, e.TaskId)
  } else {
    log.Printf("WARN Unknown task status %s\n", e.TaskStatus)
  }
}

func parseEvent(event []byte) (Event, bool) {
  var e Event
  err := json.Unmarshal(event, &e)
  return e, err == nil
}

var eventQueue = make(chan []byte)

func eventsWorker(applicationMap map[string]Application) {
  for {
    event := <-eventQueue
    e, ok := parseEvent(event)
    if ok && e.EventType == "status_update_event" {
      processStatusUpdateEvent(applicationMap, e)
      updateHAProxyConfig(applicationMap)
    }
  }
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err == nil {
    eventQueue <- body
  }
}

func startEventService(applicationMap map[string]Application) {
  go eventsWorker(applicationMap)
  http.HandleFunc("/events", eventsHandler)
  http.ListenAndServe(":8080", nil)
}
