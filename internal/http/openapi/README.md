# Documentação OpenAPI - Suassu API

## Visão Geral

Este diretório contém a implementação da documentação OpenAPI 3.1.0 para a Suassu API, seguindo a estratégia de **sem overengineering** com arquivos estáticos e embed.

## Estrutura

```
internal/http/openapi/
├── openapi.yaml      # Especificação OpenAPI 3.1.0
├── docs.html         # Interface Swagger UI
├── embed.go          # Embed dos arquivos estáticos
├── handlers.go       # Handlers para servir os arquivos
├── routes.go         # Rotas do OpenAPI
├── handlers_test.go  # Testes dos handlers
└── README.md         # Esta documentação
```

## Endpoints

### 1. Especificação OpenAPI
- **URL**: `GET /api/v1/openapi.yaml`
- **Content-Type**: `application/yaml`
- **Descrição**: Retorna a especificação OpenAPI 3.1.0 completa

### 2. Documentação Swagger UI
- **URL**: `GET /api/v1/docs`
- **Content-Type**: `text/html; charset=utf-8`
- **Descrição**: Interface interativa para testar a API

### 3. Arquivos Estáticos (Opcional)
- **URL**: `GET /api/v1/openapi/static/{arquivo}`
- **Descrição**: Serve arquivos estáticos específicos se necessário

## Características

### ✅ Implementado
- **OpenAPI 3.1.0**: Versão mais recente da especificação
- **Swagger UI**: Interface via CDN (sem dependências locais)
- **Embed**: Arquivos incluídos no binário Go
- **Sem Overengineering**: Sem libs pesadas ou geradores
- **Testes**: Cobertura completa dos handlers
- **CORS**: Configurado para permitir acesso à documentação

### 🔧 Configuração
- **Servidor**: `/api/v1` (base path)
- **Autenticação**: Bearer JWT para endpoints protegidos
- **Cache**: Configurado para arquivos estáticos (1 hora)

## Workflow de Desenvolvimento

### 1. Atualizar Especificação
```bash
# Editar o arquivo openapi.yaml
vim internal/http/openapi/openapi.yaml
```

### 2. Validar Localmente
```bash
# Instalar Redocly CLI
npm install -g @redocly/cli

# Validar a especificação
npx @redocly/cli lint internal/http/openapi/openapi.yaml
```

### 3. Testar
```bash
# Executar testes
go test ./internal/http/openapi

# Compilar e executar
go build -o suassu-api ./cmd/api
./suassu-api
```

### 4. Acessar Documentação
- **Especificação**: http://localhost:8080/api/v1/openapi.yaml
- **Swagger UI**: http://localhost:8080/api/v1/docs

## Validação em Produção

### Swagger Validator
```bash
# Validar especificação online
curl -s "https://validator.swagger.io/validator/debug?url=http://localhost:8080/api/v1/openapi.yaml"
```

### Testes de Fumaça
```bash
# Verificar especificação
curl -s http://localhost:8080/api/v1/openapi.yaml | grep "openapi: \"3.1.0\""

# Verificar Swagger UI
curl -s http://localhost:8080/api/v1/docs | grep "SwaggerUIBundle"
```

## Schemas Principais

### Autenticação
- `SignInInput`: Email e senha para login
- `SignInOutput`: Token JWT de acesso
- `UserMeResponse`: Informações do usuário autenticado

### Usuários
- `CreateUserInput`: Dados para criar usuário
- `CreateUserResponse`: ID do usuário criado
- `User`: Entidade completa do usuário
- `UserListResponse`: Lista paginada de usuários

### Erros
- `ErrorResponse`: Formato padrão de erro da API

## Segurança

- **JWT Bearer**: Autenticação via token
- **Enterprise ID**: Isolamento por empresa
- **CORS**: Configurado para permitir acesso à documentação

## Manutenção

### Atualizações
1. Modificar `openapi.yaml` conforme mudanças na API
2. Executar testes: `go test ./internal/http/openapi`
3. Recompilar: `go build -o suassu-api ./cmd/api`

### Versionamento
- A especificação é versionada junto com a API
- Sempre incluir mudanças no changelog
- Manter compatibilidade com versões anteriores quando possível

## Troubleshooting

### Problemas Comuns

#### 1. Arquivo não encontrado
```bash
# Verificar se o embed está funcionando
go test ./internal/http/openapi
```

#### 2. Swagger UI não carrega
- Verificar se `/api/v1/openapi.yaml` retorna 200
- Verificar console do navegador para erros JavaScript

#### 3. Especificação inválida
```bash
# Validar com Redocly
npx @redocly/cli lint internal/http/openapi/openapi.yaml
```

### Logs
- Verificar logs da aplicação para erros de embed
- Verificar status codes dos endpoints OpenAPI
