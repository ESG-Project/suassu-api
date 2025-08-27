# ===========================
# Suassu API — Makefile
# ===========================

GO  ?= go
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
	$(GO) run cmd/main.go

# ===========================
# section: Qualidade
# ===========================

.PHONY: lint
lint: ## Roda go vet
	$(GO) vet $(PKG)

.PHONY: tidy
tidy: ## Ajusta go.mod/go.sum
	$(GO) mod tidy

# ===========================
# section: Testes
# ===========================

.PHONY: test
test: ## Roda todos os testes
	$(GO) test $(PKG) -race

.PHONY: test-coverage
test-coverage: ## Roda todos os testes com cobertura
	$(GO) test $(PKG) -race -cover
