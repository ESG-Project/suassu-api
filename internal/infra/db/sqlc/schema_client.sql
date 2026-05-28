-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Client" (
  id text PRIMARY KEY,
  "fantasyName" text,
  "userId" text NOT NULL UNIQUE,
  FOREIGN KEY ("userId") REFERENCES "User" (id) ON DELETE CASCADE
);

