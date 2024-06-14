-- create "app_basic_auth_access_control_lists" table
CREATE TABLE "public"."app_basic_auth_access_control_lists" (
  "id" bigserial NOT NULL,
  "name" text NULL,
  "generated_name" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_app_basic_auth_access_control_lists_generated_name" UNIQUE ("generated_name")
);
-- create "app_basic_auth_access_control_users" table
CREATE TABLE "public"."app_basic_auth_access_control_users" (
  "id" bigserial NOT NULL,
  "username" text NULL,
  "encrypted_password" text NULL,
  "app_basic_auth_access_control_list_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_app_basic_auth_access_control_lists_users" FOREIGN KEY ("app_basic_auth_access_control_list_id") REFERENCES "public"."app_basic_auth_access_control_lists" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
