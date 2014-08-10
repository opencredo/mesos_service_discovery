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
  if (!ok) { t.Error("Couldn't parse event"); }

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
