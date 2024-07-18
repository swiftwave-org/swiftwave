-- modify "application_groups" table
ALTER TABLE "public"."application_groups" DROP CONSTRAINT "uni_application_groups_name", ADD COLUMN "logo" text NULL, ADD COLUMN "stack_content" text NULL;
