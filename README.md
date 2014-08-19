Marathon Service Discovery with HAProxy
=======================================

Reload HAProxy every time Marathon creates or deletes a running task. 

## Building

### Add to GOPATH

Move the `src/github.com` directory into your GOPATH and run `go install github.com/opencredo/mesos_service_discovery`

### Build script 

Running `./build.sh` should produce a binary in `$PWD/bin/`


## Usage

Start the `mesos_service_discovery` binary; optionally with the following arguments:

| Argument | Default | Description |
|----------|---------|-------------|
| host | localhost | The host Marathon is running on |
| port | 8080      | The port Marathon is running on |

Register the new service with Marathon:

```
curl -X POST http://MARATHON_HOST:MARATHON_PORT/v2/eventSubscriptions\?callbackUrl\=http://CALLBACK_HOST:8080/events
```

Replace MARATHON_HOST, MARATHON_PORT and CALLBACK_HOST (the host mesos_service_discovery is running on) 

## TODO 

* Proper logging
* Automatically (un)register with Marathon
* Separate HAProxy configuration generation / move towards a plugin based system

## License

Copyright 2014 Bart Spaans <bart.spaans@opencredo.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
