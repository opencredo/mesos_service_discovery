package main

import (
  "fmt"
  "log"
)

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


// Returns true if the applicationMap has changed
func addTask(applicationMap map[string]Application, appId string, task Task) bool {
  app, ok := applicationMap[appId]
  if !ok {
    loadExistingApps(applicationMap)
    app, ok = applicationMap[appId]
    if !ok {
      log.Printf("ERR Unknown application %s\n", appId)
      return false
    }
    return true
  }
  _, ok = app.ApplicationInstances[task.Id]
  if ok {
    return false
  }
  app.ApplicationInstances[task.Id] = task
  var port = ""
  if len(task.Ports) != 0 {
    port = fmt.Sprintf(":%d", task.Ports[0])
  }
  log.Printf("INFO Found task for %s on %s%s [%s]\n", appId, task.Host, port, task.Id)
  return true
}

func removeTask(applicationMap map[string]Application, appId string, taskId string) {
  app := applicationMap[appId]
  delete(app.ApplicationInstances, taskId)
  if len(applicationMap[appId].ApplicationInstances) == 0 {
    log.Printf("INFO Removing application '%s' from cache, because it has no running tasks left", appId)
    delete(applicationMap, appId)
  }
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


func loadExistingTasks(applicationMap map[string]Application, appId string) {
  var app MarathonTasksResponse
  err := getResponseJSON(getMarathonAddress() + "/v2/apps/" + appId, &app)
  if err != nil {
    return
  }
  for _, task := range app.App.Tasks {
    addTask(applicationMap, appId, task)
  }

  if len(applicationMap[appId].ApplicationInstances) == 0 {
    log.Printf("INFO Removing application '%s' from cache, because there are no running tasks", appId)
    delete(applicationMap, appId)
  }
}

func loadExistingApps(applicationMap map[string]Application) {
  var applications MarathonAppsResponse
  log.Printf("INFO Initializing known applications by talking to %s/v2/apps\n", getMarathonAddress())
  err := getResponseJSON(getMarathonAddress() + "/v2/apps", &applications)
  if err != nil {
    return
  }

  for _, app := range applications.Apps {
    _, ok := applicationMap[app.Id]
    if !ok {
      log.Printf("INFO Found application: %s\n", app.Id)
      app.ApplicationInstances = make(map[string]Task)
      applicationMap[app.Id] = app
      loadExistingTasks(applicationMap, app.Id)
    }
  }
}

func initApplicationMap() map[string]Application {
  applicationMap := make(map[string]Application)
  loadExistingApps(applicationMap)
  return applicationMap
}
