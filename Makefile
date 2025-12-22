DOCKER_COMPOSE_FILE=./backend/docker/docker-compose.yml

run:
	go run ./cmd/main.go

docker-up:
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

docker-ps:
	docker ps -a