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

## 📡 Endpoints (User)

### Criar usuário

```http
POST /api/v1/users
Content-Type: application/json

{
  "name": "Ana",
  "email": "ana@example.com",
  "password": "Secreta123",
  "document": "12345678900",
  "enterpriseId": "uuid-empresa"
}
```

### Listar usuários

```http
GET /api/v1/users?limit=10&offset=0
```

### Buscar por e-mail

```http
GET /api/v1/users/by-email?email=ana@example.com
```

---

## 📚 Documentação Complementar

* [📐 Arquitetura e Padrões](docs/architecture.md)
* [🏗️ Situação e Padrão do Banco de Dados](docs/database.md)
