#!/usr/bin/env bash 

export GOBIN=${PWD}/bin
export GOPATH=${PWD}

go install github.com/opencredo/mesos_service_discovery
