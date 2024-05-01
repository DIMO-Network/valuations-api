# valuations-api

There are four entry points to this application:

1. Command line Batch script `pull-valuations`. This is what is run by the kubernetes cronjob to pull new valuations every so often.
   It is defined in /cmd/valuations-api/pull_valuations.go
2. Kafka event consumer. This triggers new valuations from newly paired vehicles. Listens to events topic, and filters for
   the `com.dimo.zone.device.mint` event type. This is currently disabled b/c we were getting way too many events - most 
   likely need to filter by more parameters to isolate only new pairings that have a tokenId. 
   defined in `vehicle_mint_consumer.go`
3. REST API. Serves up some rest endpoints that Frontend clients use to get previously pulled valuations, or request a new instant offer.
   defined in `internal/api/api.go` -> `StartWebAPI`.
4. gRPC served endpoints used for internal cluster communication operations.

## Developing locally

This application has various dependencies, which can be viewed in main.go
Aside from needing a database and kafka locally, also needs to connect via gRPC to following services:
- device-definitions-api (future, move to identity-api)
- devices-api (future, move to identity-api)
- device-data-api (future, move to telemetry-api)
- users-api

Currently do not have a short & easy way to run this, but steps basically boil down to:

- get https://github.com/DIMO-Network/cluster-local and follow instructions there so that you have all above dependencies
- `$ cp settings.sample.yaml settings.yaml`
- modify `settings.yaml` so that dependent services port numbers match what is being hosted by local cluster (everything should be localhost)
- create local db if not exists: `create database valuations_api with owner dimo;`. Assumes you're using `dimo` user locally.
- run migrations: `go run ./cmd/valuations-api migrate`
- Run batch script: `go run ./cmd/valuations-api pull-valuations`
- Run events consumer, REST and gRPC: `go run ./cmd/valuations-api`

Thoughts on improving local dev:
- Only require dependencies needed for the entrypoint of the app you're trying to run (eg. batch script doesn't need kafka consumer).
- migrating to identity-api may allow some reduction in dependencies, and locally could just run against dev data in cloud.

## GRPC Generating client and server code

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

# Local development

## Swagger

Generate swagger docs:
`swag init -g cmd/valuations-api/main.go --parseDependency --parseInternal --generatedTime true`

Requirements:
`go install github.com/swaggo/swag/cmd/swag@latest`

## Migrations

`goose -dir internal/infrastructure/db/migrations create slugs_not_null sql`

Run the migrations up: `go run ./cmd/valuations-api migrate`

Regen the models: `sqlboiler psql --no-tests --wipe`