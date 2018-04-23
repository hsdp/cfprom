# cfprom

Prometheus metrics for CPU/Memory usage of apps deployed in a Cloud foundry space

## Usage

Deploy cfprom to any CF space and it will create a Prometheus `/metrics` endpoint which can be scraped. It uses the CF API to fetch statistics on all running applications. Currently it requires credentials of a CF account with the `Auditor` role or better. 

## Configuration

The following environment variables are used  

| Variable |  Required | Description |
|----------|-----------|-------------|
| CF\_API  | Y         | The Cloudfoundry API endpoint |
| CF\_USERNAME | Y     | The CF login to use |
| CF\_PASSWORD | Y     | The CF password to use |

## License

Apache. Also see the NOTICE file.
