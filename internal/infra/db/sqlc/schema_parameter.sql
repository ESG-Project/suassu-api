-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Parameter" (
  id text PRIMARY KEY,
  title text NOT NULL,
  value text,
  "enterpriseId" text NOT NULL,
  "isDefault" boolean NOT NULL DEFAULT false,
  FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise" (id)
);
