-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Product" (
  id text PRIMARY KEY,
  name text NOT NULL,
  "suggestedValue" text,
  "enterpriseId" text NOT NULL,
  "parameterId" text,
  deliverable boolean NOT NULL,
  "typeProductId" text,
  "isDefault" boolean NOT NULL DEFAULT false,
  FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise" (id),
  FOREIGN KEY ("parameterId") REFERENCES "Parameter" (id)
);
