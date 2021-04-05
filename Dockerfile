FROM golang:latest as builder
WORKDIR /go/src/app
COPY . .
ENV GOGC=off
RUN CGO_ENABLED=0 go build -o app .
ENTRYPOINT ["./app"]