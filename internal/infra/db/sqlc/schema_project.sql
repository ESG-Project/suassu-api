-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Project" (
  id text PRIMARY KEY,
  title text NOT NULL,
  cnpj text,
  activity text NOT NULL,
  codram text,
  "usefulArea" text,
  "totalArea" text,
  "pollutingPower" text,
  stage text,
  situation text,
  "addressId" text NOT NULL,
  "clientId" text NOT NULL,
  size text,
  FOREIGN KEY ("addressId") REFERENCES "Address" (id),
  FOREIGN KEY ("clientId") REFERENCES "Client" (id)
);

