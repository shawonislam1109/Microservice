CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "role_id" bigint NOT NULL,
  "name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  FOREIGN KEY ("role_id") REFERENCES "roles" ("id")
);
