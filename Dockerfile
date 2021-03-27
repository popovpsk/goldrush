FROM golang:latest as builder
WORKDIR /go/src/app
COPY . .

RUN CGO_ENABLED=0 go build -o app .

FROM scratch

ENV GOGC=1000
COPY --from=builder /go/src/app/app .
ENTRYPOINT ["./app"]
