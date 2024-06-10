# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := build

.PHONY: build clean test lint test-e2e release default license precommit go-fmt collector-builder collector-generate test-collector
default: build

release: clean test lint
	goreleaser build --clean

precommit: clean go-fmt test lint license test-e2e

build: test lint
	go build -o ./dist/ ./cmd/

clean:
	@echo "Cleaning..."
	@rm -rf ./dist ./.cache
	@rm -rf ./internal/tools/bin

OCB := internal/tools/bin/ocb
collector-builder: $(OCB)
$(OCB): internal/tools/go.mod
	cd internal/tools; \
	mkdir -p bin && true; \
	go build -o bin/ocb go.opentelemetry.io/collector/cmd/builder;

collector-generate: cmd/collector/main.go
cmd/collector/main.go: cmd/collector/builder-config.yaml $(OCB)
	cd cmd/collector; \
	../../$(OCB) --config builder-config.yaml --skip-compilation; \
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
	docker run -t --rm -v $$(pwd):/app -v ~/.cache/golangci-lint/v1.59.1:/root/.cache -w /app golangci/golangci-lint:v1.59.1 golangci-lint run -v

license:
	npx github:viperproject/check-license-header#v1 check --config .github/license-config.json

go-fmt:
	go fmt ./...
	cd test/go; go fmt ./...

go-mod-tidy:
	go mod tidy
	cd test/go; go mod tidy
	cd internal/tools; go mod tidy

update: collector-generate
	go get -u ./...
	cd test/go; go get -u ./...
	cd internal/tools; go get -u ./...
	go work sync
	make go-fmt
	make go-mod-tidy