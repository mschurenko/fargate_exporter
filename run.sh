#!/bin/bash -xe

docker run \
--rm \
--name fargate_entrypoint \
-p 2112:2112 \
fargate_entrypoint echo hi
