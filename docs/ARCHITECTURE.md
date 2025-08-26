
# Arquitetura e Padrões

Este documento descreve a arquitetura e os padrões utilizados na **Suassu API**, baseada em **Clean Architecture** e boas práticas de desenvolvimento em Go.

---

## 🧩 Camadas Principais

### 1. **Domain / Entities**
- Local: `internal/domain`
- Entidades puras do negócio (ex: `User`).
- Não dependem de frameworks, banco ou libs externas.
- Tipos simples (`string`, `*string`) para máxima compatibilidade.

### 2. **Application / Use Cases**
- Local: `internal/app`
- Contém regras de negócio e casos de uso.
- Define **ports** (interfaces) que descrevem dependências externas (`Repo`, `Hasher`).
- Implementa serviços (`Service`) que orquestram entidades e regras.

### 3. **Infrastructure / Adapters**
- Local: `internal/infra`
- Implementa os ports definidos na aplicação:
  - **Postgres + sqlc** (`user_repo.go`) → persistência de dados.
  - **bcrypt** (`bcrypt.go`) → hash de senha.
- Responsável por conversão entre tipos do banco (`sql.NullString`) e tipos do domínio (`*string`).

### 4. **Interface / Delivery**
- Local: `internal/http/v1`
- Exposição da API via HTTP (REST).
- Roteamento com **chi**.
- Handlers convertem **JSON ↔ Entities** e chamam os casos de uso.
- Rotas seguem convenção REST no plural (`/users`).

### 5. **Main (Composition Root)**
- Local: `cmd/api/main.go`
- Ponto de entrada da aplicação.
- Configura dependências (Repo, Hasher, Service).
- Conecta ao banco, carrega `.env`, sobe servidor HTTP.

---

## 🔑 Padrões Utilizados

- **Clean Architecture (Ports & Adapters)**
  O domínio não depende de detalhes. Infraestrutura depende do app, nunca o contrário.

- **Repository Pattern**
  Repositórios encapsulam a persistência.
  Interface `Repo` no app, implementação `UserRepo` no infra.

- **Dependency Injection**
  Manual no `main.go`, injetando dependências nas structs (`Service`, `Repo`).

- **Security by Design**
  Senhas nunca são salvas em texto plano.
  Hash com bcrypt → armazenado no campo `password`.

- **Null Safety**
  Uso de `*string` no domínio para opcionais.
  Conversão transparente para `sql.NullString` no repo.

- **RESTful Conventions**
  Rotas em plural (`/users`), métodos HTTP corretos (GET, POST).

---

## 🔄 Fluxo de Execução

```text
[HTTP Request] → Handler (/users)
   ↓ parse JSON
   ↓ chama Service (caso de uso)
   ↓ validações + regras de negócio
   ↓ Repo (Postgres via sqlc)
   ↓ SQL executado no banco
   ↓ retorna Entity
   ↓ serializa JSON de resposta
[HTTP Response]
````

---

## ✅ Benefícios

* **Evolutivo**: permite migração gradual do Node.js para Go.
* **Testável**: casos de uso podem ser testados isoladamente com mocks.
* **Manutenível**: responsabilidades bem separadas em camadas.
* **Flexível**: entrega HTTP pode ser trocada (gRPC, GraphQL) sem afetar domínio.
* **Seguro**: senhas protegidas com hash e nunca expostas em responses.
