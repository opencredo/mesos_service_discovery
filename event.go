package main

import (
  "fmt"
  "encoding/json"
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
		fmt.Printf("INFO Removed task for %s on %s [%s]\n", e.AppId, e.Host, e.TaskId)
	} else {
		fmt.Printf("WARN Unknown task status %s\n", e.TaskStatus)
	}
}

func parseEvent(event []byte) (Event, bool) {
	var e Event
	err := json.Unmarshal(event, &e)
	return e, err == nil
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

