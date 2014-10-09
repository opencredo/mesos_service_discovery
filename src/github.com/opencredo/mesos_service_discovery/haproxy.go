package main

import (
  "log"
  "os"
  "io"
  "io/ioutil"
  "os/exec"
  "regexp"
  "strings"
  "text/template"
)

func appExposesPorts (app Application) bool {
  return len(app.Ports) != 0;
}

func sanitizeApplicationId(appId string) string {
  return strings.Replace(appId, "/", "_", -1)
}

func getApplicationPort(app Application) int {
  return app.Ports[0]
}

func getTaskPort(task Task) int {
  return task.Ports[0]
}

func stripVersion(appId string) string {
  re := regexp.MustCompile("^(.*)-[0-9]+$")
  return re.ReplaceAllString(appId, "$1")
}

func updateHAProxyConfig(applicationMap map[string]Application, haproxyTpl string) {
  tmp, err := ioutil.TempFile("", "haproxy.cfg")
  if err != nil {
    return
  }
  generateHAProxyConfig(tmp, applicationMap, haproxyTpl)
  replaceHAProxyConfiguration(tmp.Name())
  reloadHAProxy()
}

func generateHAProxyConfig(tmp *os.File, applicationMap map[string]Application, haproxyTpl string) {
  funcMap := template.FuncMap {
    "appExposesPorts": appExposesPorts,
    "sanitizeApplicationId": sanitizeApplicationId,
    "port": getApplicationPort,
    "taskPort": getTaskPort,
    "stripVersion": stripVersion,
  }
  tpl, err := template.New("haproxy").Funcs(funcMap).Parse(haproxyTpl);
  if err != nil { panic(err); }
  err = tpl.Execute(tmp, applicationMap);
  if err != nil { panic(err); }
}

func replaceHAProxyConfiguration(tmpFile string) {
  r, err := os.Open(tmpFile);
  if (err != nil) {
    log.Printf("ERR Couldn't open %s: %s", tmpFile, err)
    return
  }
  defer r.Close()

  w, err := os.Create("/etc/haproxy/haproxy.cfg")
  if (err != nil) {
    log.Printf("ERR Couldn't open /etc/haproxy/haproxy.cfg: %s", err)
    return
  }
  defer w.Close()

  _, err = io.Copy(w, r)
  if err != nil {
    log.Printf("ERR Couldn't write /etc/haproxy/haproxy.cfg: %s", err)
    return
  }
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
