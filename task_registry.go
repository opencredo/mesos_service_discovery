package main

import (
  "fmt"
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
		fmt.Printf("ERR Unknown application %s\n", appId)
		return
	}
	app.ApplicationInstances[task.Id] = task
	fmt.Printf("INFO Found task for %s on %s:%d [%s]\n", appId, task.Host, task.Ports[0], task.Id)
}

func removeTask(applicationMap map[string]Application, appId string, taskId string) {
	app := applicationMap[appId]
	delete(app.ApplicationInstances, taskId)
}
