# Situação do Banco de Dados

Atualmente o banco de dados utilizado pela **Suassu API** (herdado do MVP em Node.js + Prisma) **não segue um padrão consistente de nomenclatura e modelagem**.

---

## 🔎 Contexto Atual
- O schema foi construído inicialmente de forma rápida para suportar o MVP.
- Há mistura de **camelCase** (`addressId`, `roleId`) e **snake_case** em alguns pontos.
- Alguns nomes de tabelas e relacionamentos não estão consistentes (singular vs plural).
- Relações e colunas podem não refletir boas práticas de modelagem.

> Em resumo: **o banco está funcional**, mas **não segue um padrão uniforme**.

---

## 📌 Decisão Atual
- O projeto em Go será **adaptado ao schema existente**, sem forçar uma padronização imediata.
- Isso garante compatibilidade com o sistema Node.js ainda em execução e evita retrabalho no curto prazo.
- Toda a camada de acesso ao banco está isolada (via **sqlc + repositórios**), permitindo evoluir o schema depois.

---

## 🔮 Plano Futuro
- Após a refatoração completa do projeto em Node.js e migração total para Go:
  - Revisar **nomenclatura de tabelas e colunas** (adotar snake_case consistente).
  - Ajustar **nomes de relações** para pluralidade clara (`users`, `projects`, etc.).
  - Documentar convenções (padrão de nomes, chaves estrangeiras, índices).
  - Criar **migrations versionadas** (via `golang-migrate` ou `atlas`) para gerenciar evolução do schema.

---

## ✅ Benefício da Abordagem
- Mantemos **compatibilidade com o sistema legado** até a transição completa.
- Evitamos inconsistências entre Node.js (Prisma) e Go (sqlc).
- Preparamos o terreno para uma **refatoração estruturada e controlada** do banco após estabilização da API em Go.

---

## 🏗️ Padrão do Banco (após padronização)

- **Nomes**: tabelas no **plural**, colunas em **snake_case**.
- **Chaves primárias**: todas as tabelas com `id UUID PRIMARY KEY`.
- **Timestamps**: `created_at` e `updated_at` em todas as tabelas.
- **Relações**: chaves estrangeiras `*_id` com índices para performance.
- **Consistência**: campos opcionais `NULL`; obrigatórios `NOT NULL`.
- **Senhas**: em `password_hash` (nunca texto puro).
- **Enums/Status**: padronizados via `ENUM` ou `CHECK`.
