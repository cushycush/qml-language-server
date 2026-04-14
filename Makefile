.PHONY: build test lint clean docker docker-run gen-grammar

BINARY_NAME=qml-language-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

# Regenerate the prebuilt grammar blob. Run after editing qmljs.grammar.json.
# Takes ~12s (grammargen is slow); the blob it produces is why startup is
# milliseconds instead of seconds.
gen-grammar:
	go run ./grammars/internal/gen

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*

docker:
	docker build -t $(BINARY_NAME):latest .

docker-run:
	docker run --rm -it $(BINARY_NAME):latest

docker-push:
	docker push $(BINARY_NAME):latest

install: build
	mkdir -p ~/.local/bin
	cp $(BINARY_NAME) ~/.local/bin/
	chmod +x ~/.local/bin/$(BINARY_NAME)

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  clean       - Remove built files"
	@echo "  docker      - Build Docker image"
	@echo "  docker-run  - Run Docker container"
	@echo "  install     - Install to ~/.local/bin"
	@echo "  coverage    - Generate coverage report"
