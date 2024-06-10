-- modify "ingress_rules" table
ALTER TABLE "public"."ingress_rules" ADD COLUMN "https_redirect" boolean NULL DEFAULT false;
