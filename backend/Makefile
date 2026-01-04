DOCKER_COMPOSE_FILE=./docker/docker-compose.yml

run:
	go run ./cmd/main.go

lint: 
	golangci-lint run ./internal/... ./cmd/...

coverage:
	go test ./internal/... ./cmd/... -coverprofile=cover.out

test:
	go test ./internal/... ./cmd/... -v

cache-clear:
	golangci-lint cache clean
	go clean -modcache

docker-up:
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

docker-ps:
	docker ps -a