#!/bin/bash -xe

GOOS=linux go build -o fargate_entrypoint main.go

docker build --no-cache -t fargate_entrypoint .
