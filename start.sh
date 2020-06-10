#!/bin/bash
# this if for deploying without docker
export GOMAXPROCS=1
export PORT="9080"
export NAME="gateway"
export SERVICE1="http://127.0.0.1:30081"
export SERVICE2="http://127.0.0.1:30082"
go run main.go &
export GOMAXPROCS=1
export PORT="30081"
export NAME="service1"
go run main.go &
export GOMAXPROCS=1
export PORT="30082"
export NAME="service2"
go run main.go &
read -r -d '' _ </dev/tty
