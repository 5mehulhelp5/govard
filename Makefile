DOCKER_ORG ?= ddtcorex/govard-
BINARY_NAME=govard
BUILD_DIR=bin
TEST_BINARY=$(BUILD_DIR)/govard-test
UNIT_PACKAGES=$(shell go list ./... | grep -v '^govard/tests/integration$$')
COVER_PACKAGES=$(shell go list ./internal/... | tr '\n' ',' | sed 's/,$$//')
VERSION_RAW ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo 1.0.0)
VERSION ?= $(patsubst v%,%,$(VERSION_RAW))
LDFLAGS ?= -s -w -X govard/internal/cmd.Version=$(VERSION)

.PHONY: build clean test test-fast test-unit test-coverage test-integration test-integration-ci test-frontend build-test-binary install lint fmt vet push test-realenv-setup test-realenv test-realenv-clean

build:
	@echo "Building Govard..."
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/govard-darwin-arm64 cmd/govard/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/govard-linux-amd64 cmd/govard/main.go

test: test-frontend test-unit test-integration

test-fast: test-frontend test-unit

test-unit:
	@echo "Running unit tests..."
	go test $(UNIT_PACKAGES) -v -short

test-coverage:
	@echo "Running unit tests with coverage..."
	go test ./tests -coverprofile=coverage.out -covermode=atomic -coverpkg=$(COVER_PACKAGES)
	go tool cover -func=coverage.out
	@echo "Coverage profile written to coverage.out"

test-integration: build-test-binary
	@echo "Running integration tests..."
	go test -tags integration ./tests/integration/... -v -timeout 30m

test-integration-ci: build-test-binary
	@echo "Running integration tests (CI mode)..."
	go test -tags integration ./tests/integration/... -v -timeout 30m -parallel 4

test-frontend:
	@echo "Running frontend unit tests..."
	node --test tests/frontend/*.test.mjs

build-test-binary:
	@echo "Building test binary..."
	mkdir -p $(BUILD_DIR)
	go build -mod=mod -ldflags "$(LDFLAGS)" -tags integration -o $(TEST_BINARY) cmd/govard/main.go

lint:
	@echo "Running linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

# Real environment tests (3 Magento 2 instances)
REALENV_DIR := tests/integration/realenv

.PHONY: test-realenv-setup test-realenv test-realenv-clean test-realenv-full

test-realenv-setup: build-test-binary
	@echo "Setting up three-environment test infrastructure..."
	@cd $(REALENV_DIR) && chmod +x setup-three-env.sh && ./setup-three-env.sh

test-realenv:
	@echo "Running real environment tests..."
	@go test -mod=mod -tags realenv ./tests/integration/realenv/... -v -timeout 30m

test-realenv-clean:
	@echo "Cleaning up real environment..."
	@cd $(REALENV_DIR) && ./setup-three-env.sh cleanup 2>/dev/null || true
	@docker-compose -f $(REALENV_DIR)/docker-compose.three-env.yml down -v 2>/dev/null || true

# Full realenv test cycle
test-realenv-full: test-realenv-clean test-realenv-setup test-realenv test-realenv-clean

install:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) cmd/govard/main.go
	sudo mv $(BINARY_NAME) /usr/local/bin/

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean -testcache

images:
	@echo "Building Govard Docker Images..."
	docker build -t $(DOCKER_ORG)apache:2.4 --build-arg APACHE_VERSION=2.4.66 docker/apache
	docker build -t $(DOCKER_ORG)apache:latest --build-arg APACHE_VERSION=2.4.66 docker/apache
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:8.15.0 --build-arg ELASTICSEARCH_VERSION=8.15.0 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:7.17.10 --build-arg ELASTICSEARCH_VERSION=7.17.10 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:7.16.3 --build-arg ELASTICSEARCH_VERSION=7.16.3 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:7.10.2 --build-arg ELASTICSEARCH_VERSION=7.10.2 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:7.9.3 --build-arg ELASTICSEARCH_VERSION=7.9.3 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:7.7.1 --build-arg ELASTICSEARCH_VERSION=7.7.1 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:7.6.2 --build-arg ELASTICSEARCH_VERSION=7.6.2 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:6.8.23 --build-arg ELASTICSEARCH_VERSION=6.8.23 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:5.6.16 --build-arg ELASTICSEARCH_VERSION=5.6.16 docker/elasticsearch
	docker build -f docker/elasticsearch/Dockerfile -t $(DOCKER_ORG)elasticsearch:2.4.6 --build-arg ELASTICSEARCH_VERSION=2.4.6 --build-arg ELASTICSEARCH_IMAGE=elasticsearch docker/elasticsearch
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:11.4 --build-arg MARIADB_VERSION=11.4 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.11 --build-arg MARIADB_VERSION=10.11 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.6 --build-arg MARIADB_VERSION=10.6 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.5 --build-arg MARIADB_VERSION=10.5 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.4 --build-arg MARIADB_VERSION=10.4 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.3 --build-arg MARIADB_VERSION=10.3 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.2 --build-arg MARIADB_VERSION=10.2 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.1 --build-arg MARIADB_VERSION=10.1 docker/mariadb
	docker build -f docker/mariadb/Dockerfile -t $(DOCKER_ORG)mariadb:10.0 --build-arg MARIADB_VERSION=10.0 docker/mariadb
	docker build -f docker/mysql/Dockerfile -t $(DOCKER_ORG)mysql:8.4 --build-arg MYSQL_VERSION=8.4 docker/mysql
	docker build -f docker/mysql/Dockerfile -t $(DOCKER_ORG)mysql:8.0 --build-arg MYSQL_VERSION=8.0 docker/mysql
	docker build -f docker/mysql/Dockerfile -t $(DOCKER_ORG)mysql:5.7 --build-arg MYSQL_VERSION=5.7 docker/mysql
	docker build -t $(DOCKER_ORG)nginx:1.28 --build-arg NGINX_VERSION=1.28.0 docker/nginx
	docker build -t $(DOCKER_ORG)nginx:latest --build-arg NGINX_VERSION=1.28.0 docker/nginx
	docker build -f docker/opensearch/Dockerfile -t $(DOCKER_ORG)opensearch:3.0.0 --build-arg OPENSEARCH_VERSION=3.0.0 docker/opensearch
	docker build -f docker/opensearch/Dockerfile -t $(DOCKER_ORG)opensearch:2.19.0 --build-arg OPENSEARCH_VERSION=2.19.0 docker/opensearch
	docker build -f docker/opensearch/Dockerfile -t $(DOCKER_ORG)opensearch:2.12.0 --build-arg OPENSEARCH_VERSION=2.12.0 docker/opensearch
	docker build -f docker/opensearch/Dockerfile -t $(DOCKER_ORG)opensearch:2.5.0 --build-arg OPENSEARCH_VERSION=2.5.0 docker/opensearch
	docker build -f docker/opensearch/Dockerfile -t $(DOCKER_ORG)opensearch:1.3.20 --build-arg OPENSEARCH_VERSION=1.3.20 docker/opensearch
	docker build -f docker/opensearch/Dockerfile -t $(DOCKER_ORG)opensearch:1.2.0 --build-arg OPENSEARCH_VERSION=1.2.0 docker/opensearch
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:8.4 --build-arg PHP_VERSION=8.4 docker/php
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:8.3 --build-arg PHP_VERSION=8.3 docker/php
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:8.2 --build-arg PHP_VERSION=8.2 docker/php
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:8.1 --build-arg PHP_VERSION=8.1 docker/php
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:7.4 --build-arg PHP_VERSION=7.4 docker/php
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:7.3 --build-arg PHP_VERSION=7.3 docker/php
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:7.2 --build-arg PHP_VERSION=7.2 docker/php
	docker build -f docker/php/Dockerfile -t $(DOCKER_ORG)php:7.1 --build-arg PHP_VERSION=7.1 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:8.4 --build-arg PHP_VERSION=8.4 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:8.3 --build-arg PHP_VERSION=8.3 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:8.2 --build-arg PHP_VERSION=8.2 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:8.1 --build-arg PHP_VERSION=8.1 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:7.4 --build-arg PHP_VERSION=7.4 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:7.3 --build-arg PHP_VERSION=7.3 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:7.2 --build-arg PHP_VERSION=7.2 docker/php
	docker build -f docker/php/magento2/Dockerfile -t $(DOCKER_ORG)php-magento2:7.1 --build-arg PHP_VERSION=7.1 docker/php
	docker build -f docker/rabbitmq/Dockerfile -t $(DOCKER_ORG)rabbitmq:4.1 --build-arg RABBITMQ_VERSION=4.1 docker/rabbitmq
	docker build -f docker/rabbitmq/Dockerfile -t $(DOCKER_ORG)rabbitmq:3.13 --build-arg RABBITMQ_VERSION=3.13 docker/rabbitmq
	docker build -f docker/rabbitmq/Dockerfile -t $(DOCKER_ORG)rabbitmq:3.12 --build-arg RABBITMQ_VERSION=3.12 docker/rabbitmq
	docker build -f docker/rabbitmq/Dockerfile -t $(DOCKER_ORG)rabbitmq:3.11 --build-arg RABBITMQ_VERSION=3.11 docker/rabbitmq
	docker build -f docker/rabbitmq/Dockerfile -t $(DOCKER_ORG)rabbitmq:3.9 --build-arg RABBITMQ_VERSION=3.9 docker/rabbitmq
	docker build -f docker/rabbitmq/Dockerfile -t $(DOCKER_ORG)rabbitmq:3.8 --build-arg RABBITMQ_VERSION=3.8 docker/rabbitmq
	docker build -f docker/rabbitmq/Dockerfile -t $(DOCKER_ORG)rabbitmq:3.7 --build-arg RABBITMQ_VERSION=3.7 docker/rabbitmq
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:7.4 --build-arg REDIS_VERSION=7.4 docker/redis
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:7.2 --build-arg REDIS_VERSION=7.2 docker/redis
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:7.0 --build-arg REDIS_VERSION=7.0 docker/redis
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:6.2 --build-arg REDIS_VERSION=6.2 docker/redis
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:6.0 --build-arg REDIS_VERSION=6.0 docker/redis
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:5.0 --build-arg REDIS_VERSION=5.0 docker/redis
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:4.0 --build-arg REDIS_VERSION=4.0 docker/redis
	docker build -f docker/redis/Dockerfile -t $(DOCKER_ORG)redis:3.2 --build-arg REDIS_VERSION=3.2 docker/redis
	docker build -f docker/valkey/Dockerfile -t $(DOCKER_ORG)valkey:8.0 --build-arg VALKEY_VERSION=8.0 docker/valkey
	docker build -f docker/valkey/Dockerfile -t $(DOCKER_ORG)valkey:7.2 --build-arg VALKEY_VERSION=7.2 docker/valkey
	docker build -t $(DOCKER_ORG)varnish:7.6 --build-arg VARNISH_VERSION=7.6 docker/varnish
	docker build -t $(DOCKER_ORG)varnish:7.4 --build-arg VARNISH_VERSION=7.4 docker/varnish
	docker build -t $(DOCKER_ORG)varnish:7.0 --build-arg VARNISH_VERSION=7.0 docker/varnish
	docker build -t $(DOCKER_ORG)varnish:6.0 --build-arg VARNISH_VERSION=6.0 docker/varnish
	docker build -t $(DOCKER_ORG)varnish:latest --build-arg VARNISH_VERSION=7.6 docker/varnish

push:
	@echo "Pushing Govard Docker Images..."
	docker push $(DOCKER_ORG)apache:2.4
	docker push $(DOCKER_ORG)apache:latest
	docker push $(DOCKER_ORG)elasticsearch:8.15.0
	docker push $(DOCKER_ORG)elasticsearch:7.17.10
	docker push $(DOCKER_ORG)elasticsearch:7.16.3
	docker push $(DOCKER_ORG)elasticsearch:7.10.2
	docker push $(DOCKER_ORG)elasticsearch:7.9.3
	docker push $(DOCKER_ORG)elasticsearch:7.7.1
	docker push $(DOCKER_ORG)elasticsearch:7.6.2
	docker push $(DOCKER_ORG)elasticsearch:6.8.23
	docker push $(DOCKER_ORG)elasticsearch:5.6.16
	docker push $(DOCKER_ORG)elasticsearch:2.4.6
	docker push $(DOCKER_ORG)mariadb:11.4
	docker push $(DOCKER_ORG)mariadb:10.11
	docker push $(DOCKER_ORG)mariadb:10.6
	docker push $(DOCKER_ORG)mariadb:10.5
	docker push $(DOCKER_ORG)mariadb:10.4
	docker push $(DOCKER_ORG)mariadb:10.3
	docker push $(DOCKER_ORG)mariadb:10.2
	docker push $(DOCKER_ORG)mariadb:10.1
	docker push $(DOCKER_ORG)mariadb:10.0
	docker push $(DOCKER_ORG)mysql:8.4
	docker push $(DOCKER_ORG)mysql:8.0
	docker push $(DOCKER_ORG)mysql:5.7
	docker push $(DOCKER_ORG)nginx:1.28
	docker push $(DOCKER_ORG)nginx:latest
	docker push $(DOCKER_ORG)opensearch:3.0.0
	docker push $(DOCKER_ORG)opensearch:2.19.0
	docker push $(DOCKER_ORG)opensearch:2.12.0
	docker push $(DOCKER_ORG)opensearch:2.5.0
	docker push $(DOCKER_ORG)opensearch:1.3.20
	docker push $(DOCKER_ORG)opensearch:1.2.0
	docker push $(DOCKER_ORG)php:8.4
	docker push $(DOCKER_ORG)php:8.3
	docker push $(DOCKER_ORG)php:8.2
	docker push $(DOCKER_ORG)php:8.1
	docker push $(DOCKER_ORG)php:7.4
	docker push $(DOCKER_ORG)php:7.3
	docker push $(DOCKER_ORG)php:7.2
	docker push $(DOCKER_ORG)php:7.1
	docker push $(DOCKER_ORG)php-magento2:8.4
	docker push $(DOCKER_ORG)php-magento2:8.3
	docker push $(DOCKER_ORG)php-magento2:8.2
	docker push $(DOCKER_ORG)php-magento2:8.1
	docker push $(DOCKER_ORG)php-magento2:7.4
	docker push $(DOCKER_ORG)php-magento2:7.3
	docker push $(DOCKER_ORG)php-magento2:7.2
	docker push $(DOCKER_ORG)php-magento2:7.1
	docker push $(DOCKER_ORG)rabbitmq:4.1
	docker push $(DOCKER_ORG)rabbitmq:3.13
	docker push $(DOCKER_ORG)rabbitmq:3.12
	docker push $(DOCKER_ORG)rabbitmq:3.11
	docker push $(DOCKER_ORG)rabbitmq:3.9
	docker push $(DOCKER_ORG)rabbitmq:3.8
	docker push $(DOCKER_ORG)rabbitmq:3.7
	docker push $(DOCKER_ORG)redis:7.4
	docker push $(DOCKER_ORG)redis:7.2
	docker push $(DOCKER_ORG)redis:7.0
	docker push $(DOCKER_ORG)redis:6.2
	docker push $(DOCKER_ORG)redis:6.0
	docker push $(DOCKER_ORG)redis:5.0
	docker push $(DOCKER_ORG)redis:4.0
	docker push $(DOCKER_ORG)redis:3.2
	docker push $(DOCKER_ORG)valkey:8.0
	docker push $(DOCKER_ORG)valkey:7.2
	docker push $(DOCKER_ORG)varnish:7.6
	docker push $(DOCKER_ORG)varnish:7.4
	docker push $(DOCKER_ORG)varnish:7.0
	docker push $(DOCKER_ORG)varnish:6.0
	docker push $(DOCKER_ORG)varnish:latest
