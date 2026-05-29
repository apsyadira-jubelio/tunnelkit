.PHONY: build run migrate clean

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

migrate:
	./scripts/migrate.sh

clean:
	rm -rf bin/

test:
	go test ./...

lint:
	golangci-lint run
