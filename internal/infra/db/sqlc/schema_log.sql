-- Apenas para o sqlc entender tipos (não roda no banco).
-- "tag" é o id do usuário que originou o evento (nome herdado do Prisma).
CREATE TABLE "Log" (
  id text PRIMARY KEY,
  tag text,
  "enterpriseId" text NOT NULL,
  description text NOT NULL,
  "createdAt" timestamp NOT NULL DEFAULT now(),
  FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise" (id),
  FOREIGN KEY (tag) REFERENCES "User" (id)
);
