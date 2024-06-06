-- modify "users" table
ALTER TABLE "public"."users" ADD COLUMN "totp_enabled" boolean NULL DEFAULT false, ADD COLUMN "totp_secret" text NULL;
