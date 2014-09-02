package main

import (
  "fmt"
  "log"
  "os"
  "io/ioutil"
  "os/exec"
  "strings"
)

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

func updateHAProxyConfig(applicationMap map[string]Application) {
  tmp, err := ioutil.TempFile("", "haproxy.cfg")
  if err != nil {
    return
  }
  generateHAProxyConfig(tmp, applicationMap)
  replaceHAProxyConfiguration(tmp.Name())
  reloadHAProxy()
}

func generateHAProxyConfig(tmp *os.File, applicationMap map[string]Application) {
  fmt.Fprintf(tmp, haproxyHeader)
  for appId, app := range applicationMap {
    var safeAppId = strings.Replace(appId, "/", "_", -1)
    fmt.Fprintf(tmp, "\nlisten %s\n  bind 0.0.0.0:%d\n  mode tcp\n  option tcplog\n  balance leastconn\n", safeAppId, app.Ports[0])
    i := 0
    for _, task := range app.ApplicationInstances {
      fmt.Fprintf(tmp, "  server %s-%d %s:%d check\n", safeAppId, i, task.Host, task.Ports[0])
      i++
    }
  }
}

func replaceHAProxyConfiguration(tmpFile string) {
  err := os.Rename(tmpFile, "/etc/haproxy/haproxy.cfg")
  if err != nil {
    log.Printf("ERR Couldn't write /etc/haproxy/haproxy.cfg: %s", err)
    return
  }
  log.Println("INFO Written new /etc/haproxy/haproxy.cfg")
}

func reloadHAProxy() {
  cmd := exec.Command("service", "haproxy", "reload")
  err := cmd.Start()
  if err != nil {
    log.Println("ERR failed to reload HAProxy")
    return
  }
  err = cmd.Wait()
  if err != nil {
    log.Println("ERR failed to reload HAProxy")
    return
  }
}
