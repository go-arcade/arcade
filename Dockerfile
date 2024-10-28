FROM alpine:latest

RUN mkdir -p /opt/arcade/bin /opt/arcade/conf.d

COPY ./arcade /opt/arcade/bin

EXPOSE 8080

WORKDIR /opt/arcade

ENTRYPOINT [ "./bin/arcade -conf conf.d/config.toml" ]
