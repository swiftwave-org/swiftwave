-- reverse: modify "applications" table
ALTER TABLE "public"."applications" DROP CONSTRAINT "fk_application_groups_applications", DROP COLUMN "application_group_id";
-- reverse: create "application_groups" table
DROP TABLE "public"."application_groups";
-- reverse: modify "applications" table
ALTER TABLE "public"."applications" ADD COLUMN "application_group" text NULL;