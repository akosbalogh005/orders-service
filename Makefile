# Variables
APP_NAME ?= orders-service
BIN_DIR ?= bin
BIN_PATH ?= $(BIN_DIR)/$(APP_NAME)
DOCKER_IMAGE ?= $(APP_NAME):latest

.PHONY: build docker-build test

build:
	@echo ">> Building binary at $(BIN_PATH)"
	@mkdir -p $(BIN_DIR)
	GO111MODULE=on go build -o $(BIN_PATH) ./cmd/server

docker-build:
	@echo ">> Building Docker image $(DOCKER_IMAGE)"
	docker build -t $(DOCKER_IMAGE) .

test:
	@echo ">> Running unit tests"
	go test ./...

