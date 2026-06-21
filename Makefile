BIN     := bin/server
OLD_MOD := go-template

.PHONY: run build migrate test lint tidy rename

run:
	go run ./cmd

build:
	go build -o $(BIN) ./cmd

migrate:
	go run ./cmd/migrate

test:
	go test ./...

lint:
	go vet ./...

tidy:
	go mod tidy

# Usage: make rename MODULE=github.com/you/myapp
rename:
	@test -n "$(MODULE)" || (echo "usage: make rename MODULE=github.com/you/myapp"; exit 1)
	find . -name "*.go" -not -path "./.git/*" | xargs sed -i 's|$(OLD_MOD)|$(MODULE)|g'
	sed -i 's|$(OLD_MOD)|$(MODULE)|g' go.mod
