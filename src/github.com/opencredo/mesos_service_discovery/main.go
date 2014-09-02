package main

import (
  "encoding/json"
  "flag"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
)

var LOCAL_HOST    = flag.String("host", "localhost", "The host this service runs on.")
var LOCAL_PORT    = flag.String("port", "8081", "The port to run this service on")
var MARATHON_HOST = flag.String("marathon_host", "localhost", "The host Marathon is running on")
var MARATHON_PORT = flag.String("marathon_port", "8080", "The port Marathon is running on")

func getMarathonAddress() string {
  return "http://" + *MARATHON_HOST + ":" + *MARATHON_PORT
}

func getThisServiceAddress() string {
  return "http://" + *LOCAL_HOST + ":" + *LOCAL_PORT + "/events"
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

func registerWithMarathon() {
  log.Printf("INFO Registering this service (%s) with Marathon", getThisServiceAddress())
  
  var postAddress = getMarathonAddress() + "/v2/eventSubscriptions"
  var urlParams = make(url.Values)
  urlParams.Add("callbackUrl", getThisServiceAddress())
  resp, err := http.Post(postAddress + "?" + urlParams.Encode(), "application/json", nil)
  if err != nil {
    log.Fatal("FATAL Couldn't register service with Marathon")
  }
  log.Println("INFO Successfully registered with Marathon")
  resp.Body.Close()
}

func main() {
  log.Println("INFO Application started")
  flag.Parse()
  applicationMap := initApplicationMap()
  registerWithMarathon()
  startEventService(applicationMap, *LOCAL_PORT)
}
