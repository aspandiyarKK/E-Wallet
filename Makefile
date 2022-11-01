
lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run ./...

up:
	docker-compose up -d db

run:
	docker-compose up -d

down:
	docker-compose down

test:up
	go test -v ./...
	make down