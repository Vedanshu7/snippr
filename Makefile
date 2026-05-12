.PHONY: run build build-cli build-ui install test lint docker-build docker-up docker-down

run:
	go run ./cmd/server

build: build-ui
	mkdir -p bin
	go build -o bin/snippr-server ./cmd/server
	go build -o bin/snippr ./cmd/snippr

build-ui:
	cd web && npm run build
	rm -rf cmd/server/ui
	cp -r web/dist cmd/server/ui

build-cli:
	mkdir -p bin
	go build -o bin/snippr ./cmd/snippr

install:
	go install ./cmd/snippr

test:
	go test ./...

lint:
	go vet ./...

docker-build:
	docker build -t snippr .

docker-up:
	docker compose up -d

docker-down:
	docker compose down
