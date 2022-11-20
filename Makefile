.PHONY: migrate migrate_down migrate_up migrate_version docker_build docker_start 

# ==============================================================================
# Go migrate postgresql

force:
	migrate -database postgres://postgres:postgres@localhost:5432/vessels_db?sslmode=disable -path migrations force 1

version:
	migrate -database postgres://postgres:postgres@localhost:5432/vessels_db?sslmode=disable -path migrations version

migrate_up:
	migrate -database postgres://postgres:postgres@localhost:5432/vessels_db?sslmode=disable -path migrations up 

migrate_down:
	migrate -database postgres://postgres:postgres@localhost:5432/vessels_db?sslmode=disable -path migrations down 


# ==============================================================================
# Docker compose commands

docker_build:
	echo "Starting local environment"
	docker-compose -f docker-compose.yaml up --build

docker_start:
	echo "Starting local environment"
	docker-compose up -d 


# ==============================================================================
# Tools commands

run-linter:
	echo "Starting linters"
	golangci-lint run ./...


# ==============================================================================
# Main

run:
	go run ./cmd/main.go

build:
	go build ./cmd/main.go

test:
	go test -cover ./...


# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache


# ==============================================================================
# Docker support

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

clean:
	docker system prune -f

logs-local:
	docker logs -f $(FILES)