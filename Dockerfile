FROM alpine:latest

COPY fargate_exporter /usr/local/bin/fargate_exporter

ENTRYPOINT [ "/usr/local/bin/fargate_exporter" ]
