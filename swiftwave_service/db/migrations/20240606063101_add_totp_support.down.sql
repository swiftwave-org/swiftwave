-- reverse: modify "users" table
ALTER TABLE "public"."users" DROP COLUMN "totp_secret", DROP COLUMN "totp_enabled";
