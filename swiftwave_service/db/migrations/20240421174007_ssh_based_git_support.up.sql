-- modify "deployments" table
ALTER TABLE "public"."deployments" ADD COLUMN "git_type" text NULL DEFAULT 'http', ADD COLUMN "git_endpoint" text NULL, ADD COLUMN "git_ssh_user" text NULL;
-- modify "git_credentials" table
ALTER TABLE "public"."git_credentials" ADD COLUMN "type" text NULL DEFAULT 'http', ADD COLUMN "ssh_private_key" text NULL, ADD COLUMN "ssh_public_key" text NULL;
-- patch "deployments" table
UPDATE "public"."deployments" SET "git_type" = 'http', "git_endpoint" = 'github.com' WHERE "git_provider" = 'github';
UPDATE "public"."deployments" SET "git_type" = 'http', "git_endpoint" = 'gitlab.com' WHERE "git_provider" = 'gitlab';