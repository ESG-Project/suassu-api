# ===========================
# Suassu API — Makefile
# ===========================

PKG ?= ./...

.PHONY: help
help: ## Mostra esta ajuda
	@echo ""
	@echo "\033[1mSuassu API — Comandos disponíveis:\033[0m"
	@echo ""
	@grep -E '^(# section: |[a-zA-Z0-9_.-]+:.*##)' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS=":.*##"} \
	/^\# section: / { title=$$0; sub(/^\# section: /,"",title); printf "\n# %s\n", title; next } \
	/^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }'
	@echo ""


# ===========================
# section: Executar
# ===========================
.PHONY: run
run: ## Executa o projeto
	go run cmd/api/main.go

.PHONY: sqlc
sqlc: ## Gera código sqlc
	sqlc generate

# ===========================
# section: Qualidade
# ===========================

.PHONY: lint
lint: ## Roda go vet
	go vet $(PKG)

.PHONY: tidy
tidy: ## Ajusta go.mod/go.sum
	go mod tidy

# ===========================
# section: Testes
# ===========================

.PHONY: test
test: ## Roda todos os testes
	go test $(PKG) -race

.PHONY: test-coverage
test-coverage: ## Roda todos os testes com cobertura
	go test $(PKG) -race -cover

# ===========================
# section: Docker
# ===========================

.PHONY: docker-build
docker-build: ## Constrói a imagem Docker
	docker build -t suassu-api .

.PHONY: docker-build-nocache
docker-build-nocache: ## Constrói a imagem Docker sem cache
	docker build --no-cache -t suassu-api .

.PHONY: docker-up
docker-up: ## Inicia apenas a aplicação Go
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Para a aplicação Go
	docker-compose down

.PHONY: docker-logs
docker-logs: ## Mostra logs da aplicação
	docker-compose logs -f api

.PHONY: docker-stats
docker-stats: ## Mostra stats da aplicação
	docker stats

.PHONY: docker-restart
docker-restart: ## Reinicia a aplicação
	docker-compose restart api

.PHONY: docker-clean
docker-clean: ## Remove containers, imagens e volumes não utilizados
	docker system prune -f
	docker volume prune -f
