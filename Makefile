help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[32m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## Build image
	docker compose build

up: ## Up containers
	docker compose up -d --remove-orphans

down: ## Down containers
	docker compose down --remove-orphans

logs: ## Show logs
	docker compose logs

logsf: ## Follow logs
	docker compose logs -f

migrate: ## Execute migrations
	docker compose run --rm goose

in-client:
	docker compose exec client cmd/client/client --addr=server:50051 user.db p4ssw0rd
