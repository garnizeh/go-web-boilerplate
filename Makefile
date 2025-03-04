APP_NAME ?= app

# ==============================================================================
# Install external tools

install-tools: install-staticcheck install-migrate install-sqlc install-expvarmon install-MailHog install-tailwindcss

.PHONY: install-staticcheck
install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: install-migrate
install-migrate:
	go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: install-sqlc
install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: install-expvarmon
install-expvarmon:
	go install github.com/divan/expvarmon@latest

.PHONY: install-mailhog
install-mailhog:
	go install github.com/mailhog/MailHog@latest

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
	goose -dir storage/datastore/sql/migrations -s create $(filter-out $@,$(MAKECMDGOALS)) sql

generate:
	sqlc generate

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

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: test
test:
	go test -race -v -timeout 30s ./...

# ==============================================================================
# Building views

.PHONY: tailwind-build
tailwind-build:
	tailwindcss -i ./static/css/input.css -o ./static/css/style.min.css --minify

.PHONY: templ-generate
templ-generate:
	templ generate

# ==============================================================================
# Building app

dev:
	go build -o ./tmp/main ./cmd/main.go && air

.PHONY: dependencies
dependencies:
	go mod tidy
	go mod vendor

build: vet staticcheck dependencies tailwind-build templ-generate test
	go build -ldflags="-s -w" -o ./bin/$(APP_NAME) ./cmd/app/main.go

# ==============================================================================
# Metrics and Tracing

metrics:
	expvarmon -ports="localhost:3010" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

statsviz:
	open "Google Chrome" http://localhost:3010/debug/statsviz

# ==============================================================================
# SMTP support

smtp:
	MailHog

# ==============================================================================
# Docker support

.PHONY: docker-build
docker-build:
	docker-compose -f ./dev/docker-compose.yml build

.PHONY: docker-up
docker-up:
	docker-compose -f ./dev/docker-compose.yml up

.PHONY: docker-dev
docker-dev:
	docker-compose -f ./dev/docker-compose.yml -f ./dev/docker-compose.dev.yml up

.PHONY: docker-down
docker-down:
	docker-compose -f ./dev/docker-compose.yml down

.PHONY: docker-clean
docker-clean:
	docker-compose -f ./dev/docker-compose.yml down -v --rmi all