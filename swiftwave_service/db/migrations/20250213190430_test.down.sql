-- reverse: create index "idx_user_sessions_session_id" to table: "user_sessions"
DROP INDEX "public"."idx_user_sessions_session_id";
-- reverse: create "user_sessions" table
DROP TABLE "public"."user_sessions";
