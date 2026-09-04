-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "TypeProduct" (
  id text PRIMARY KEY,
  type text NOT NULL,
  "enterpriseId" text NOT NULL,
  FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise" (id) ON DELETE CASCADE
);
