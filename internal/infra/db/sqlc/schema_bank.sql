-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Bank" (
  id text PRIMARY KEY,
  code text NOT NULL UNIQUE,
  name text NOT NULL UNIQUE
);

CREATE TABLE "EnterpriseBank" (
  id text PRIMARY KEY,
  "enterpriseId" text NOT NULL,
  "bankId" text NOT NULL,
  FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise" (id) ON DELETE CASCADE,
  FOREIGN KEY ("bankId") REFERENCES "Bank" (id) ON DELETE CASCADE
);
