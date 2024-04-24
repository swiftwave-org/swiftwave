-- reverse: modify "ingress_rules" table
ALTER TABLE "public"."ingress_rules" DROP COLUMN "external_service", DROP COLUMN "target_type";
