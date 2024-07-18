-- reverse: modify "application_groups" table
ALTER TABLE "public"."application_groups" DROP COLUMN "stack_content", DROP COLUMN "logo", ADD CONSTRAINT "uni_application_groups_name" UNIQUE ("name");
