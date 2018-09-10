# cfprom

Prometheus metrics for CPU/Memory usage of apps deployed in a Cloud foundry space

# Building

## Requirements

- [Go 1.11 or newer](https://golang.org/doc/install)

## Compiling

Clone the repo somewhere (preferably outside your GOPATH):

```
$ git clone git@github.com:hsdp/cfprom
$ cd cfprom
$ go build .
```

This produce a logproxy binary exectable read for use

# Docker

Alternatively, you can use the included Dockerfile to build a docker image which can be deployed to CF directly.

```
$ git clone git@github.com:hsdp/cfprom
$ cd cfprom
$ docker build -t cfprom .
```

## Usage

Deploy cfprom to any CF space and it will create a Prometheus `/metrics` endpoint which can be scraped. It uses the CF API to fetch statistics on all running applications. Currently it requires credentials of a CF account with the `Auditor` role or better. 

## Configuration

The following environment variables are used  

| Variable |  Required | Description |
|----------|-----------|-------------|
| CF\_USERNAME | N     | The CF login to use |
| CF\_PASSWORD | N     | The CF password to use |

## Bootstrapping

If you do not wish to add `CF_USERNAME` and/or `CF_PASSWORD` to the environment you can bootstrap cfprom by posting the username and password to the `/bootstrap` endpoint:

```
curl -X POST https://cfprom.<your_cf_domain>/bootstrap -d '{"username":"admin","password":"SuperS3cret"}'
```

Only after sending the correct credentials will cfprom be able to start collecting metrics. Note that this a tradeoff between security and convenience. You will have to bootstrap again if cfprom gets restarted or restaged for whatever reason.

## License

Apache. Also see the NOTICE file.
