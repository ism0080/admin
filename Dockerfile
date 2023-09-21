# syntax=docker/dockerfile:1
FROM golang:1.21.0 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /pocketbase

FROM gcr.io/distroless/base-debian11 AS release

WORKDIR /

COPY --from=build /pocketbase /pocketbase

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/pocketbase"]
CMD ["serve", "--http=0.0.0.0:8080"]