#!/bin/bash

docker run --rm -ti \
-v /home/ec2-user/signalfx/config:/config \
quay.io/signalfx/signalfx-agent:4.6.3 \
signalfx-agent -config /config/config.yaml
