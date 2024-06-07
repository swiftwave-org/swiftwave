-- create "config_mounts" table
CREATE TABLE "public"."config_mounts" (
  "id" bigserial NOT NULL,
  "application_id" text NULL,
  "config_id" text NULL,
  "content" text NULL,
  "mounting_path" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_applications_config_mounts" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
