fmt:
	go fmt
.PHONY: fmt

lint: lint
	golangci-lint run
.PHONY: lint

vet: fmt
	go vet
.PHONY: vet

build: vet
	go mod tidy
	go build -buildvcs=false
.PHONY: build
