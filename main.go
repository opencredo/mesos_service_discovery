package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "net/http"
)

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

func main() {
  flag.Parse()
  fmt.Printf("Running things and stuff\n")
  applicationMap := initApplicationMap()
  startEventService(applicationMap)
}
