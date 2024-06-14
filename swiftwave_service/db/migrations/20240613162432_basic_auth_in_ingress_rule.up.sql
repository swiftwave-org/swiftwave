-- modify "ingress_rules" table
ALTER TABLE "public"."ingress_rules" ADD COLUMN "authentication_auth_type" text NULL DEFAULT 'none', ADD COLUMN "authentication_app_basic_auth_access_control_list_id" bigint NULL, ADD
 CONSTRAINT "fk_app_basic_auth_access_control_lists_ingress_rules" FOREIGN KEY ("authentication_app_basic_auth_access_control_list_id") REFERENCES "public"."app_basic_auth_access_control_lists" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
