BRANCH ?= main
BUILD_N ?= 0

PORT ?= 6554
DB_URL ?= postgres://user:password@localhost:5432/accounts?sslmode=disable

build:
	@rm -rf bin/*
	go build -ldflags="-X 'main.Version=1.0.0.$(BUILD_N)-$(BRANCH)'" -o ./bin/accounts ./cmd/accounts

run: 
	DB_URL=$(DB_URL) PORT=$(PORT) go run cmd/accounts/main.go