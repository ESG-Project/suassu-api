# Arquitetura e Padrões

Este documento descreve a arquitetura e os padrões utilizados na **Suassu API**, baseada em **Clean Architecture** e boas práticas de desenvolvimento em Go.

---

## 🧩 Camadas Principais

### 1. **Domain / Entities**

* Local: `internal/domain`
* Entidades puras do negócio (ex.: `User`).
* Não dependem de frameworks, banco ou libs externas.
* Tipos simples (`string`, `*string`) para máxima compatibilidade.
* Define também erros de domínio (`ErrInvalidUser`, etc).

### 2. **Application / Use Cases**

* Local: `internal/app`
* Contém regras de negócio e casos de uso.
* Define **ports** (interfaces) que descrevem dependências externas (`Repo`, `Hasher`, `TokenIssuer`).
* Implementa serviços (`Service`) que orquestram entidades e regras.
* Multi-tenant: todos os casos de uso privados recebem `enterpriseID` como parâmetro explícito.

### 3. **Infrastructure / Adapters**

* Local: `internal/infra`
* Implementa os ports definidos na aplicação:

  * **Postgres + sqlc** (`user_repo.go`) → persistência de dados.
  * **bcrypt** (`bcrypt.go`) → hash de senha.
  * Logging com **zap**.
* Responsável por conversão entre tipos do banco (`sql.NullString`) e tipos do domínio (`*string`).

### 4. **Interface / Delivery**

* Local: `internal/http`
* Subpacotes por módulo (`authhttp`, `userhttp`) e middlewares (`httpmw`).
* Exposição da API via HTTP (REST) com **chi**, prefixada em `/api/v1`.
* Handlers convertem **JSON ↔ Entities** e chamam os casos de uso.
* Rotas seguem convenção REST no plural (`/users`).
* Middleware JWT (`AuthJWT`) injeta claims no contexto.
* Middleware `RequireEnterprise` opcional garante `enterpriseId` nas rotas privadas.

### 5. **Config**

* Local: `internal/config`
* Carrega variáveis de ambiente a partir de `.env`.
* Centraliza parâmetros de banco (DB\_DSN), JWT (secret, issuer, audience, TTL) e logging (log level).

### 6. **Error Handling**

* `internal/apperr`: define erros tipados com códigos (`invalid`, `not_found`, `unauthorized`, etc.).
* `internal/http/httperr`: converte erros em respostas HTTP JSON padronizadas, incluindo `requestId`.
* Middleware de erro garante logging estruturado e status HTTP corretos.

### 7. **Main (Composition Root)**

* Local: `cmd/api/main.go`
* Ponto de entrada da aplicação.
* Configura dependências (Repo, Hasher, Service, Logger).
* Conecta ao banco, carrega `.env`, sobe servidor HTTP.

---

## 🔑 Padrões Utilizados

* **Clean Architecture (Ports & Adapters)**
  O domínio não depende de detalhes. Infraestrutura depende do app, nunca o contrário.

* **Repository Pattern**
  Repositórios encapsulam a persistência.
  Interface `Repo` no app, implementação `UserRepo` no infra.

* **Dependency Injection**
  Manual no `main.go`, injetando dependências nas structs (`Service`, `Repo`, `Hasher`).

* **Error Handling Centralizado**
  Todo erro tratado via `apperr` + `httperr`, com responses JSON padronizadas.

* **Security by Design**

  * Senhas nunca são salvas em texto plano.
  * Hash com bcrypt → armazenado no campo `password_hash`.
  * JWT curto (ex.: 15m).
  * `enterpriseId` obrigatório em queries privadas.
  * Refresh/logout planejados em fase 2.

* **Null Safety**
  Uso de `*string` no domínio para opcionais.
  Conversão transparente para `sql.NullString` no repo.

* **RESTful Conventions**
  Rotas em plural (`/users`), versão prefixada `/api/v1`.
  Métodos HTTP corretos (GET, POST, PUT, DELETE).

* **Logging Estruturado**
  `zap` configurado no `main`.
  `requestId` via middleware do chi incluso no log e na resposta.

---

## 🔄 Fluxo de Execução

```text
[HTTP Request] → Middleware JWT/Auth → Handler (/users)
   ↓ parse JSON
   ↓ chama Service (caso de uso)
   ↓ validações + regras de negócio
   ↓ Repo (Postgres via sqlc) com enterpriseId
   ↓ SQL executado no banco
   ↓ retorna Entity
   ↓ serializa JSON de resposta
[HTTP Response] (com requestId)
```

---

## ✅ Benefícios

* **Evolutivo**: permite migração gradual do Node.js para Go.
* **Testável**: casos de uso testados isoladamente com mocks.
* **Manutenível**: responsabilidades bem separadas em camadas.
* **Flexível**: entrega HTTP pode ser trocada (gRPC, GraphQL) sem afetar domínio.
* **Seguro**: senhas com hash bcrypt, JWT curto, multi-tenant garantido.
* **Observável**: logging estruturado + requestId em todas as respostas.
