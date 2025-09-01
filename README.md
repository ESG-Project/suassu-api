# Suassu API

API construída em **Go** utilizando **Clean Architecture**.
Este projeto é a refatoração gradual de um MVP que originalmente foi desenvolvido em Node.js com Prisma e PostgreSQL.

---

## 🚀 Tecnologias

- **Go 1.22+**
- **PostgreSQL**
- **sqlc** → geração de código a partir de SQL
- **Chi** → roteador HTTP
- **Zap** → logging estruturado
- **bcrypt** → hash de senha

---

## 📂 Estrutura de Pastas (resumida)

```text
cmd/api/                 → main.go (entrypoint da aplicação)
internal/
  app/                   → casos de uso e regras de negócio
  domain/                → entidades do domínio
  http/v1/               → camada HTTP (handlers e rotas)
  infra/                 → adapters (Postgres, bcrypt, sqlc)
.env                     → variáveis de ambiente
.env.example             → exemplo de configuração
sqlc.yaml                → config do sqlc
````

---

## ⚙️ Configuração

1. **Variáveis de ambiente**

   Copie o `.env.example` para `.env` e configure:

   ```env
   APP_NAME=suassu-api
   APP_ENV=dev
   HTTP_PORT=8080

   DB_DSN=
   DATABASE_URL=postgres://usuario:senha@localhost:5432/nome_db?sslmode=disable
   ```

   > A aplicação usa `DB_DSN`. Se estiver vazio, usa `DATABASE_URL`.

2. **Geração de código sqlc**

   O arquivo `sqlc.yaml` já está configurado.
   Gere o código para que os pacotes em `internal/infra/db/sqlc/gen` sejam criados.

3. **Rodar aplicação**

   ```bash
   go run ./cmd/api
   ```

   O servidor sobe na porta definida em `HTTP_PORT` (padrão: `8080`).

---

## 📡 API Endpoints

**📚 Documentação completa da API está disponível no Swagger UI:**
- **URL**: `http://localhost:8080/api/v1/docs`
- **Especificação OpenAPI**: `http://localhost:8080/api/v1/openapi.yaml`

### Principais Funcionalidades

- **Autenticação**: Login JWT com refresh token
- **Usuários**: CRUD completo com paginação por cursor
- **Multi-tenant**: Isolamento por empresa
- **Validação**: Tratamento robusto de erros

---

## 📚 Documentação Complementar

* [📐 Arquitetura e Padrões](docs/ARCHITECTURE.md)
* [🏗️ Situação e Padrão do Banco de Dados](docs/DATABASE.md)
