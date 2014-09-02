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
| host          | localhost | The hostname to advertise to Marathon (ie. the host this service is running on) |
| port          | 8081      | The port to run this service on |
| marathon_host | localhost | The host Marathon is running on |
| marathon_port | 8080      | The port Marathon is running on |

The service will automatically try to register itself with Marathon on start-up.

## TODO 

* Automatically deregister the service on shutdown
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
