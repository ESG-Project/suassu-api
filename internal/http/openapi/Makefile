.PHONY: help test validate build clean

# Variáveis
OPENAPI_FILE = openapi.yaml
REDOCLY_CMD = npx @redocly/cli

help: ## Mostra esta ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Executa os testes do OpenAPI
	go test -v .

test-coverage: ## Executa testes com cobertura
	go test -coverprofile=coverage.out .
	go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório de cobertura gerado: coverage.html"

validate: ## Valida a especificação OpenAPI com Redocly
	@echo "Validando especificação OpenAPI..."
	@if command -v $(REDOCLY_CMD) >/dev/null 2>&1; then \
		$(REDOCLY_CMD) lint $(OPENAPI_FILE); \
	else \
		echo "Redocly CLI não encontrado. Instalando..."; \
		npm install -g @redocly/cli; \
		$(REDOCLY_CMD) lint $(OPENAPI_FILE); \
	fi

validate-online: ## Valida a especificação online (requer servidor rodando)
	@echo "Validando especificação online..."
	@if [ -z "$(PORT)" ]; then \
		echo "Defina a variável PORT (ex: make validate-online PORT=8080)"; \
		exit 1; \
	fi
	curl -s "https://validator.swagger.io/validator/debug?url=http://localhost:$(PORT)/api/v1/openapi.yaml" | jq '.'

smoke-test: ## Testes de fumaça (requer servidor rodando)
	@echo "Executando testes de fumaça..."
	@if [ -z "$(PORT)" ]; then \
		echo "Defina a variável PORT (ex: make smoke-test PORT=8080)"; \
		exit 1; \
	fi
	@echo "1. Testando especificação OpenAPI..."
	@curl -s -o /dev/null -w "Status: %{http_code}, Content-Type: %{content_type}\n" http://localhost:$(PORT)/api/v1/openapi.yaml
	@echo "2. Testando Swagger UI..."
	@curl -s -o /dev/null -w "Status: %{http_code}, Content-Type: %{content_type}\n" http://localhost:$(PORT)/api/v1/docs
	@echo "3. Verificando versão OpenAPI..."
	@curl -s http://localhost:$(PORT)/api/v1/openapi.yaml | grep -q "openapi: \"3.1.0\"" && echo "✓ Versão OpenAPI 3.1.0 confirmada" || echo "✗ Versão OpenAPI incorreta"
	@echo "4. Verificando Swagger UI..."
	@curl -s http://localhost:$(PORT)/api/v1/docs | grep -q "SwaggerUIBundle" && echo "✓ Swagger UI carregado" || echo "✗ Swagger UI não encontrado"

build: ## Compila a aplicação
	go build -o ../../../suassu-api ./cmd/api

run: ## Executa a aplicação
	@if [ -z "$(PORT)" ]; then \
		echo "Defina a variável PORT (ex: make run PORT=8080)"; \
		exit 1; \
	fi
	PORT=$(PORT) ./suassu-api

dev: ## Modo desenvolvimento (compila e executa)
	@if [ -z "$(PORT)" ]; then \
		echo "Defina a variável PORT (ex: make dev PORT=8080)"; \
		exit 1; \
	fi
	$(MAKE) build
	$(MAKE) run PORT=$(PORT)

clean: ## Remove arquivos gerados
	rm -f ../../../suassu-api
	rm -f coverage.out coverage.html

install-redocly: ## Instala o Redocly CLI
	npm install -g @redocly/cli

check-deps: ## Verifica dependências necessárias
	@echo "Verificando dependências..."
	@command -v go >/dev/null 2>&1 || { echo "✗ Go não encontrado"; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "✗ NPM não encontrado"; exit 1; }
	@echo "✓ Dependências OK"

all: check-deps validate test ## Executa validação e testes
	@echo "✅ Todas as verificações passaram!"

.PHONY: help test validate build clean run dev check-deps all
