-- create "user_sessions" table
CREATE TABLE "public"."user_sessions" (
  "id" bigserial NOT NULL,
  "user_id" bigint NULL,
  "session_id" text NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_sessions" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create index "idx_user_sessions_session_id" to table: "user_sessions"
CREATE INDEX "idx_user_sessions_session_id" ON "public"."user_sessions" ("session_id");
