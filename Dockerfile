FROM golang:alpine3.11 AS builder

WORKDIR /go/src/baquiax.me/dd-distributed-tracing

COPY . . 

RUN go build -o /tmp/app cmd/app/main.go


FROM alpine:3.11

WORKDIR /app 

COPY --from=builder /tmp/app .

ENTRYPOINT [ "./app" ]