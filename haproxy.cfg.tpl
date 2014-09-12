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

{{ range $appId, $app := . }}
{{ if appExposesPorts $app }}
listen {{ sanitizeApplicationId $appId }}
  bind 0.0.0.0:{{ port $app }}
  mode tcp
  option tcplog
  balance leastconn
  {{ range $taskId, $task := $app.ApplicationInstances }}
  server {{$taskId}} {{$task.Host}}:{{taskPort $task}} check
  {{ end }}
{{ end }}
{{ end }}
