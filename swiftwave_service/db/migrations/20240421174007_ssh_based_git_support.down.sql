-- reverse: modify "git_credentials" table
ALTER TABLE "public"."git_credentials" DROP COLUMN "ssh_public_key", DROP COLUMN "ssh_private_key", DROP COLUMN "type";
-- reverse: modify "deployments" table
ALTER TABLE "public"."deployments" DROP COLUMN "git_ssh_user", DROP COLUMN "git_endpoint", DROP COLUMN "git_type";
