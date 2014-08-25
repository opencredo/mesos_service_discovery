package main

import (
  "testing"
)

var test_event_json = `
{
  "eventType": "status_update_event",
  "taskStatus": "TASK_KILLED",
  "appId": "helloworld",
  "taskId": "test-1234",
  "host": "localhost",
  "ports": [12345]
}
`

func TestParseEvent(t *testing.T) {
  event, ok := parseEvent([]byte(test_event_json))
  if !ok {
    t.Error("Couldn't parse event")
  }

  if event.EventType != "status_update_event" {
    t.Error("Didn't parse eventType correctly")
  }
  if event.TaskStatus != "TASK_KILLED" {
    t.Error("Didn't parse taskStatus correctly")
  }
  if event.AppId != "helloworld" {
    t.Error("Didn't parse appId correctly")
  }
  if event.TaskId != "test-1234" {
    t.Error("Didn't parse taskId correctly")
  }
  if event.Host != "localhost" {
    t.Error("Didn't parse host correctly")
  }
}

func TestRemoveTask(t *testing.T) {
  m := make(map[string]Application)
  tasks := make(map[string]Task)
  task := Task{"test-task-id", "localhost", []int{}}
  tasks["test-task-id"] = task
  app := Application{"test-app-id", []int{}, tasks}
  m["test-app-id"] = app

  removeTask(m, "test-app-id", "test-task-id")

  _, exists := m["test-app-id"]
  if !exists {
    t.Error("Application shouldn't have been removed")
  }

  _, exists = m["test-app-id"].ApplicationInstances["test-task-id"]
  if exists {
    t.Error("Task should have been removed")
  }
}
