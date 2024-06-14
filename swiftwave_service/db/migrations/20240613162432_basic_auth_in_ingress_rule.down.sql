-- reverse: modify "ingress_rules" table
ALTER TABLE "public"."ingress_rules" DROP CONSTRAINT "fk_app_basic_auth_access_control_lists_ingress_rules", DROP COLUMN "authentication_app_basic_auth_access_control_list_id", DROP COLUMN "authentication_auth_type";
