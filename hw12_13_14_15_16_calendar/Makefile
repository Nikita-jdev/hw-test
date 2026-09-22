BIN := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_STORER := "./bin/calendar_storer"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build: build-calendar build-scheduler build-storer

build-calendar:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

build-scheduler:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/scheduler

build-storer:
	go build -v -o $(BIN_STORER) -ldflags "$(LDFLAGS)" ./cmd/storer

run: build
	$(BIN) -config ./configs/config.toml

run-scheduler: build-scheduler
	$(BIN_SCHEDULER) -config ./configs/config.toml

run-storer: build-storer
	$(BIN_STORER) -config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/... ./...

COMPOSE_TEST := docker compose -f docker-compose.yaml -f docker-compose.integration.yaml

integration-tests:
	cd deployments && \
	$(COMPOSE_TEST) down -v ; \
	$(COMPOSE_TEST) up --build --abort-on-container-exit --exit-code-from tests tests ; \
	status=$$? ; \
	$(COMPOSE_TEST) down -v ; \
	exit $$status

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

lint: install-lint-deps
	go build ./...
	golangci-lint run ./...

up:
	cd deployments && docker compose up -d --build

down:
	cd deployments && docker compose down -v

.PHONY: build build-calendar build-scheduler build-storer run run-scheduler run-storer build-img run-img version test lint up down integration-tests
