.DEFAULT_GOAL := build

.PHONY: fmt vet build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test

cover: test
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

build: test
	go build -o application

clean:
	go clean
	rm ./application
