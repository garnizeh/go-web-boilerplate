APP_NAME ?= boilerplate
BUILD := $(shell git rev-parse HEAD)
GOENVPATH = $(shell go env GOPATH)

# ==============================================================================
# Install external tools

.PHONY: install-tailwindcss
install-tailwindcss:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	sudo mv tailwindcss-linux-x64 /usr/local/bin/


# ==============================================================================
# Database support

# need to pass the name of the file as argument
# ex.: make migrate-create add_users    ---->   will create a storage/datastore/sql/migrations/000X_add_users.sql file
migrate-create:
	go tool goose -dir storage/datastore/sql/migrations -s create $(filter-out $@,$(MAKECMDGOALS)) sql

generate:
	go tool sqlc generate

dev-clean:
	rm tmp/data/*

# ==============================================================================
# Git management

git-clean:
	git checkout main
	git remote update origin --prune
	git branch | grep -v "\smain\b" | xargs git branch -D

# ==============================================================================
# Checking source code

check-cicd: lint vet staticcheck sec vuln

.PHONY: lint
lint:
	go tool golangci-lint run --modules-download-mode vendor --timeout=10m -E gosec -E prealloc -E misspell -E unconvert -E goimports -E sqlclosecheck -E bodyclose -E noctx -E govet -E gosimple -E gofmt -E unparam

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	go tool staticcheck ./...

.PHONY: sec
sec:
	go tool gosec -exclude-generated ./...

.PHONY: vuln
vuln:
	@echo go tool govulncheck ./...

.PHONY: test
test:
	go test -race -v -timeout 30s ./...

# ==============================================================================
# Building views

.PHONY: templ-generate
templ-generate:
	go tool templ generate

# ==============================================================================
# Building app

all: dependencies templ-generate check-cicd test build

dev:
	go build -o ./tmp/main ./cmd/main.go && air

.PHONY: dependencies
dependencies:
	go mod tidy
	go mod vendor

build:
	go build -ldflags='-s -w -X "main.build=$(BUILD)" -X "main.appName=$(APP_NAME)"' -o ./bin/$(APP_NAME) ./cmd/app/main.go

build-prod:
	go build -ldflags='-s -w -X "main.build=$(BUILD)" -X "main.appName=$(APP_NAME)" -extldflags "-static"' -o app ./cmd/app/main.go

# ==============================================================================
# Metrics and Tracing

metrics:
	go tool expvarmon -ports="localhost:3010" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

statsviz:
	open "Google Chrome" http://localhost:3010/debug/statsviz

# ==============================================================================
# SMTP support

smtp:
	go tool MailHog

# ==============================================================================
# Docker support

.PHONY: docker-prod
docker-build:
	docker build -f ./docker/Dockerfile.prod -t $(APP_NAME):test .
	docker run --rm -p 3000:3000 -p 3010:3010 $(APP_NAME):test
