BINARY := one-ftpserver

.PHONY: build run fmt lint test check ci clean

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

fmt:
	goimports -w .

lint:
	golangci-lint run ./...

test:
	go test ./...

check: fmt lint test

ci: lint test

clean:
	rm -f $(BINARY)
