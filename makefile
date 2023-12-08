# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := build

.PHONY: build clean test lint test-e2e release default
default: build

release: clean test lint
	goreleaser build --clean

build: test lint
	go build -o ./dist/ ./cmd/

clean:
	@echo "Cleaning..."
	@rm -rf ./dist ./.cache

test:
	go test -timeout 60s ./...

test-e2e:
	cd test; \
	go test -timeout 60s -v ./...

lint:
	docker run -t --rm -v $$(pwd):/app -v ~/.cache/golangci-lint/v1.55.2:/root/.cache -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v

license:
	npx github:viperproject/check-license-header#v1 check --config .github/license-config.json