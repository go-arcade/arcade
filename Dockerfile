FROM golang:1.20 AS builder

# 设置工作目录
WORKDIR /app

COPY ./ /app/

RUN make -f build/Makefile all


FROM alpine:latest

RUN apk add --no-cache tzdata unzip&& \
mkdir -p /opt/arcade/bin /opt/arcade/conf.d

COPY --from=builder /app/arcade /opt/arcade/bin

EXPOSE 8080

WORKDIR /opt/arcade

ENTRYPOINT [ "./bin/arcade -conf conf.d/config.toml" ]
