-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Enterprise" (
  id text PRIMARY KEY,
  cnpj text NOT NULL UNIQUE,
  "addressId" text,
  email text NOT NULL,
  "fantasyName" text,
  name text NOT NULL,
  phone text,
  FOREIGN KEY ("addressId") REFERENCES "Address" (id)
);
