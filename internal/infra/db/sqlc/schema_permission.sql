CREATE TABLE "Permission" (
  id text PRIMARY KEY,
  "featureId" text NOT NULL,
  "roleId" text NOT NULL,
  "create" boolean NOT NULL DEFAULT false,
  "read" boolean NOT NULL DEFAULT false,
  "update" boolean NOT NULL DEFAULT false,
  "delete" boolean NOT NULL DEFAULT false,
  FOREIGN KEY ("featureId") REFERENCES "Feature" (id) ON DELETE CASCADE,
  FOREIGN KEY ("roleId") REFERENCES "Role" (id) ON DELETE CASCADE,
  UNIQUE ("featureId", "roleId")
);
