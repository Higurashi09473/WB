DOCKER_COMPOSE_FILE=./docker/docker-compose.yml

run:
	go run ./cmd/main.go

lint: 
	golangci-lint run ./internal/... ./cmd/...

cache-clear:
	golangci-lint cache clear

docker-up:
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

docker-ps:
	docker ps -a