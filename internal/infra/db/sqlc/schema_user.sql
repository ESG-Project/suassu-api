-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE "User" (
  id text PRIMARY KEY,
  name text NOT NULL,
  email text NOT NULL UNIQUE,
  password text NOT NULL,
  document text NOT NULL,
  phone text,
  "addressId" text,
  "roleId" text,
  "enterpriseId" text NOT NULL,
  is_super_admin boolean NOT NULL DEFAULT false,
  FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise" (id),
  FOREIGN KEY ("addressId") REFERENCES "Address" (id),
  FOREIGN KEY ("roleId") REFERENCES "Role" (id)
);
