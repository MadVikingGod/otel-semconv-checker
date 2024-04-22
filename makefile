# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := build

.PHONY: build clean test lint test-e2e release default license precommit go-fmt
default: build

release: clean test lint
	goreleaser build --clean

precommit: clean go-fmt test lint license test-e2e

build: test lint
	go build -o ./dist/ ./cmd/

clean:
	@echo "Cleaning..."
	@rm -rf ./dist ./.cache

collector-generate: cmd/collector/main.go
cmd/collector/main.go: cmd/collector/builder-config.yaml
	cd cmd/collector; \
	ocb --config builder-config.yaml --skip-compilation; \
	go mod tidy

test:
	go test -timeout 60s ./...

test-collector:
	cd cmd/collector; \
	go test -timeout 60s -v ./...

# TODO: make the Prerequisites every directory under test/ when this get bigger than 3.
test-e2e: test-e2e/go

test-e2e/go:
	cd test/go; \
	go test -timeout 60s -v ./...

# TODO: split this out for linting the server and the e2e tests.
lint:
	docker run -t --rm -v $$(pwd):/app -v ~/.cache/golangci-lint/v1.57.2:/root/.cache -w /app golangci/golangci-lint:v1.57.2 golangci-lint run -v

license:
	npx github:viperproject/check-license-header#v1 check --config .github/license-config.json

go-fmt:
	go fmt ./...
	cd test/go
	go fmt ./...