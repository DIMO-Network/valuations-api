# valuations-api

Api for managing vehicle signal decoding on the DIMO platform.

## Developing locally

**TL;DR**

```bash
cp settings.sample.yaml settings.yaml
docker compose up -d
go run ./cmd/valuations-api
```

## Generating client and server code

1. Install the protocol compiler plugins for Go using the following commands

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

2. Run protoc in the root directory

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/grpc/*.proto
```

## Linting

`brew install golangci-lint`

`golangci-lint run`

This should use the settings from `.golangci.yml`, which you can override.

If brew version does not work, download from https://github.com/golangci/golangci-lint/releases (darwin arm64 if M1), then copy to /usr/local/bin and sudo xattr -c golangci-lint

_Make sure you're running the docker image (ie. docker compose up)_

If you get a command not found error with sqlboiler, make sure your go install is correct.
[Instructions here](https://jimkang.medium.com/install-go-on-mac-with-homebrew-5fa421fc55f5)

## Local development

Importing data: Device definition exports are [here]([url](https://drive.google.com/drive/u/1/folders/1WymEqZo-bCH2Zw-m5L9u_ynMSwPeEARL))
You can use sqlboiler to import or this command:
```sh
psql "host=localhost port=5432 dbname=valuations_api user=dimo password=dimo" -c "\COPY valuations_api.integrations (id, type, style, vendor, created_at, updated_at, refresh_limit_secs, metadata) FROM '/Users/aenglish/Downloads/drive-download-20221020T172636Z-001/integrations.csv' DELIMITER ',' CSV HEADER"
```

### Starting Kafka locally

`$ brew services start kafka`
`$ brew services start zookeeper`

This will use the brew services to start kafka locally on port 9092. One nice thing of this vs. docker-compose is that we can use this 
same instance for all our different locally running services that require kafka. 

### Produce some test messages

`$ go run ./cmd/test-producer`

In current state this only produces a single message, but should be good enough starting point to test locally. 

### Create decoding topic 

`kafka-topics --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic topic.dbc.decoding`

### Sample read messages in the topic

`kafka-console-consumer --bootstrap-server localhost:9092 --topic topic.dbc.decoding --from-beginning`