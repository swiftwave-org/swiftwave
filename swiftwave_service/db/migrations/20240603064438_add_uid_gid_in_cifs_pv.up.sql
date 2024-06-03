-- modify "persistent_volumes" table
ALTER TABLE "public"."persistent_volumes" ADD COLUMN "cifs_config_uid" bigint NULL DEFAULT 0, ADD COLUMN "cifs_config_gid" bigint NULL DEFAULT 0;
