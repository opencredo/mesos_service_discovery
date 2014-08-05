#!/usr/bin/env ruby
require 'rubygems'
require 'json'
require 'net/http'

url = URI.parse("http://172.16.5.10:8080/v2/tasks")
req = Net::HTTP::Get.new(url.path)
req.add_field("Accept", "application/json")
res = Net::HTTP.new(url.host, url.port).start do |http|
  http.request(req)
end
data = res.body
lb = Hash.new { |hash, key| hash[key] = [] }
result = JSON.parse(data)
result['tasks'].each { |task|
  key = task['appId']

  value = "#{task['host']}" + ":" + "#{task['ports'].first}@@@@#{task['id']}"
  if lb.has_key?(key)
    lb[key] << value
  else
    lb[key] = [value]
  end


}

print <<"EOF";

global
  daemon
  log 127.0.0.1 local0
  log 127.0.0.1 local1 notice
  maxconn 4096

defaults
  log         global
  retries     3
  maxconn     2000
  contimeout  500
  clitimeout  500
  srvtimeout  500
  mode	      http
 errorfile 503 /etc/haproxy/errors/503.http

listen stats
  bind 0.0.0.0:9090
  balance
  mode http
  stats enable
  stats auth admin:admin

  backend default
   server Local 192.168.1.5:80 check

  backend mesos
   server mesos1 172.16.5.10:5050 check
   server mesos2 172.16.5.11:5050 check
   server mesos3 172.16.5.12:5050 check

  backend marathon
   server marathon1 172.16.5.10:8080 check
   server marathon2 172.16.5.11:8080 check
   server marathon3 172.16.5.12:8080 check

frontend http-in
    bind *:80
    acl is_mesos hdr_end(host) -i mesos.example.com
    use_backend mesos if is_mesos
    acl is_marathon hdr_end(host) -i marathon.example.com

    use_backend marathon if is_marathon
EOF

lb.each { |app|
  puts "acl is_#{app.first} hdr_end(host) -i #{app.first}.example.com"
  puts "use_backend #{app.first} if is_#{app.first}"
  puts "default_backend default"
}
lb.each { |app|

  print <<"EOF"

backend #{app.first}
  mode http
  option tcplog
  option httpchk GET /
  balance leastconn
EOF
  app.last.each {|server|
    server_split = server.split("@@@@")
    id = server_split.last
    hostAndPort= server_split.first
   puts "server #{id} #{hostAndPort} check"

  }
  puts " "
}
