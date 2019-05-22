#!/bin/bash -xe

GOOS=linux go build -o fargate_exporter main.go

tag=mschurenko/fargate_exporter

go test github.com/mschurenko/fargate_exporter/utils

docker build --no-cache -t $tag .

docker push $tag


