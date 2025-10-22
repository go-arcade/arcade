FROM golang:1.24.4 AS builder

WORKDIR /app

COPY ./ /app/

RUN apt-get install -y unzip && \
    make build

FROM FROM gcr.io/distroless/static:nonroot

WORKDIR /

RUN mkdir -p /conf.d

COPY --from=builder /app/arcade /arcade

EXPOSE 8080

ENTRYPOINT [ "/arcade -conf /conf.d/config.toml" ]
