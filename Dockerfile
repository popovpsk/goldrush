FROM golang:latest as builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go build -o app .

FROM scratch

COPY --from=builder /go/src/app/app .
ENTRYPOINT ["./app"]
