# build stage
FROM golang:1.11.0-alpine3.8 AS builder
RUN apk add --no-cache git openssh gcc musl-dev
WORKDIR /cfprom
COPY . /cfprom
RUN cd /cfprom && go build -o cfprom

FROM alpine:latest 
MAINTAINER Andy Lo-A-Foe <andy.loafoe@aemain.com>
WORKDIR /app
COPY --from=builder /cfprom/cfprom /app
RUN apk --no-cache add ca-certificates

EXPOSE 8080
CMD ["/app/cfprom"]
