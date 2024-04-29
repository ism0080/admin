# syntax=docker/dockerfile:1
FROM golang:1.21.0 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /pocketbase

FROM litestream/litestream as runner

FROM debian:stable-slim AS release
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY --from=build /pocketbase /pocketbase
COPY --from=runner /usr/local/bin/litestream /litestream

COPY etc/litestream.yml /etc/litestream.yml
COPY scripts/run.sh /scripts/run.sh

EXPOSE 8080

RUN chmod +x /scripts/run.sh

CMD ["/scripts/run.sh"]