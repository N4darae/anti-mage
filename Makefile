.PHONY: all fmt vet test check tidy clean

all: check

fmt:
	gofmt -l -w .

vet:
	go vet ./...

test:
	go test ./...

check: fmt vet test

tidy:
	go mod tidy

clean:
	rm -rf bin
