
FROM golang:1.17 AS build
RUN mkdir /app
ADD . /app/
WORKDIR /app
# COPY go.mod go.sum ./
RUN go mod download
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .
CMD ["/app/http-monitor"]

FROM alpine:3.8
RUN mkdir -p /app
COPY --from=build /app/http-monitor /app/http-monitor
COPY --from=build /app/monitor.yml /app/monitor.yml
WORKDIR /app
ENTRYPOINT ["/app/http-monitor"]