CREATE TABLE "Role" (
  id text PRIMARY KEY,
  title text NOT NULL,
  "enterpriseId" text NOT NULL,
  FOREIGN KEY ("enterpriseId") REFERENCES "Enterprise" (id) ON DELETE CASCADE,
  UNIQUE (title, "enterpriseId")
);
