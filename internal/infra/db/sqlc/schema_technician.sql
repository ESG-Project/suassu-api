-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "Technician" (
  id text PRIMARY KEY,
  "proRegister" text,
  graduation text,
  ctf text,
  "userId" text NOT NULL UNIQUE,
  FOREIGN KEY ("userId") REFERENCES "User" (id) ON DELETE CASCADE
);
