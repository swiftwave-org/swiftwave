-- create "application_groups" table
CREATE TABLE "public"."application_groups" (
  "id" text NOT NULL,
  "name" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_application_groups_name" UNIQUE ("name")
);
-- modify "applications" table
ALTER TABLE "public"."applications" ADD COLUMN "application_group_id" text NULL, ADD
 CONSTRAINT "fk_application_groups_applications" FOREIGN KEY ("application_group_id") REFERENCES "public"."application_groups" ("id") ON UPDATE CASCADE ON DELETE SET NULL;
-- create records in "application_groups" table
INSERT INTO "public"."application_groups" (id, name)
SELECT
    gen_random_uuid() AS "id",
    "application_group" AS "name"
FROM (
         SELECT DISTINCT "application_group"
         FROM "public"."applications"
         WHERE "application_group" IS NOT NULL AND "application_group" != ''
     ) AS "distinct_groups";
-- update "applications" table with newly created records in "application_groups" table
UPDATE "public"."applications"
SET application_group_id = groups.id
FROM (
         SELECT id, name
         FROM "public"."application_groups"
     ) AS groups
WHERE applications.application_group = groups.name;
-- modify "applications" table
ALTER TABLE "public"."applications" DROP COLUMN "application_group";
