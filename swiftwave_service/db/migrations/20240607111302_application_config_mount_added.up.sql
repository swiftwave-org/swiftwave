-- create "config_mounts" table
CREATE TABLE "public"."config_mounts" (
  "id" bigserial NOT NULL,
  "application_id" text NULL,
  "config_id" text NULL,
  "content" text NULL,
  "mounting_path" text NULL,
  "uid" bigint NULL DEFAULT 0,
  "gid" bigint NULL DEFAULT 0,
  "file_mode" bigint NULL DEFAULT 444,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_applications_config_mounts" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
