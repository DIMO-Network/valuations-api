# valuations-api

Api for obtaining and querying vehicle valuations. 

## Developing locally

**TL;DR**

```bash
cp settings.sample.yaml settings.yaml
docker compose up -d
go run ./cmd/valuations-api
```

Create DB locally:
`create database valuations_api with owner dimo;`

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