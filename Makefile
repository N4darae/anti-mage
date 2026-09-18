.PHONY: all fmt vet test check build tidy clean

all: check

fmt:
	gofmt -l -w .

vet:
	go vet ./...

test:
	go test ./...

check: fmt vet test

build:
	go build -o bin/anti-mage .

tidy:
	go mod tidy

clean:
	rm -rf bin
