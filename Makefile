lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run ./...

up:
	docker-compose up -d

down:
	docker-compose down

test:
	go clean -testcache
	go test -v ./...