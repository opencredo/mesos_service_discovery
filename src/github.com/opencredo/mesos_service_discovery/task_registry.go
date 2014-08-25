package main

import (
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


func addTask(applicationMap map[string]Application, appId string, task Task) {
  app, ok := applicationMap[appId]
  if !ok {
    loadExistingApps(applicationMap)
  }
  app, ok = applicationMap[appId]
  if !ok {
    log.Printf("ERR Unknown application %s\n", appId)
    return
  }
  app.ApplicationInstances[task.Id] = task
  log.Printf("INFO Found task for %s on %s:%d [%s]\n", appId, task.Host, task.Ports[0], task.Id)
}

func removeTask(applicationMap map[string]Application, appId string, taskId string) {
  app := applicationMap[appId]
  delete(app.ApplicationInstances, taskId)
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
  updateHAProxyConfig(applicationMap)
  return applicationMap
}
