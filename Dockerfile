FROM ubuntu:18.04

COPY fargate_entrypoint /usr/local/bin/fargate_entrypoint

ENTRYPOINT [ "/usr/local/bin/fargate_entrypoint" ]
