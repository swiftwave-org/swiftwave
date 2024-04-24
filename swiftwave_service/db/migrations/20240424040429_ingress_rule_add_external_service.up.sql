-- modify "ingress_rules" table
ALTER TABLE "public"."ingress_rules" ADD COLUMN "target_type" text NULL DEFAULT 'application', ADD COLUMN "external_service" text NULL;
