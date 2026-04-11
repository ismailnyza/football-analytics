APP_TUI=simtui
APP_CLI=simcli

.PHONY: build test fmt lint run-tui run-cli

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -w $(shell find cmd internal -name '*.go' -type f)

lint:
	go vet ./...

run-tui:
	go run ./cmd/simtui

run-cli:
	go run ./cmd/simcli --help

