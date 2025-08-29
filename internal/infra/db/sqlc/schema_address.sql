-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Address" (
  id text PRIMARY KEY,
  "zipCode" text NOT NULL,
  state text NOT NULL,
  city text NOT NULL,
  neighborhood text NOT NULL,
  street text NOT NULL,
  num text NOT NULL,
  latitude text,
  longitude text,
  "addInfo" text
);
