.PHONY: check vet lint test test-nocgo build run package
check: vet lint test test-nocgo build
vet:
	GOWORK=off go tool mygo vet .
lint:
	GOWORK=off go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
test:
	GOWORK=off go test -race ./...
test-nocgo:
	GOWORK=off CGO_ENABLED=0 go test ./...
build:
	GOWORK=off go build ./...
run:
	go run .
package:
	GOWORK=off go tool mygo build
