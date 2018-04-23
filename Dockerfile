# build stage
FROM golang:alpine AS builder
WORKDIR /go/src/github.com/hsdp/cfprom
COPY . .

RUN cd /go/src/github.com/hsdp/cfprom && go build -o cfprom

FROM alpine:latest 
MAINTAINER Andy Lo-A-Foe <andy.loafoe@aemain.com>
WORKDIR /app
COPY --from=builder /go/src/github.com/hsdp/cfprom/cfprom /app
RUN apk --no-cache add ca-certificates

EXPOSE 8080
CMD ["/app/cfprom"]
