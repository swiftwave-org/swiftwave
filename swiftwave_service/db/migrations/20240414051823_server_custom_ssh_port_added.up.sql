-- modify "servers" table
ALTER TABLE "public"."servers" ADD COLUMN "ssh_port" bigint NULL DEFAULT 22;
