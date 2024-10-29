FROM golang:1.23 AS builder

# 设置工作目录
WORKDIR /app

COPY ./ /app/

RUN apt-get update && \
apt-get install -y unzip && \
touch ./internal/engine/router/static/arcade.js && \
make -f build/Makefile build

FROM alpine:latest

RUN apk add --no-cache tzdata && \
mkdir -p /opt/arcade/bin /opt/arcade/conf.d

COPY --from=builder /app/arcade /opt/arcade/bin

EXPOSE 8080

WORKDIR /opt/arcade

ENTRYPOINT [ "./bin/arcade -conf conf.d/config.toml" ]
