# build stage
FROM golang:1.13.6-alpine3.10 AS builder
RUN apk add --no-cache git openssh gcc musl-dev
WORKDIR /cfprom
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build
COPY . .
RUN go build -o cfprom

FROM alpine:latest 
MAINTAINER Andy Lo-A-Foe <andy.lo-a-foe@philips.com>
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /cfprom/cfprom /app
RUN apk --no-cache add ca-certificates

EXPOSE 8080
CMD ["/app/cfprom"]
