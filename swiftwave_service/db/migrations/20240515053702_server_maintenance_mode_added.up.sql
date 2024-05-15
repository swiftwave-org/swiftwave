-- modify "servers" table
ALTER TABLE "public"."servers" ADD COLUMN "maintenance_mode" boolean NULL DEFAULT false;
