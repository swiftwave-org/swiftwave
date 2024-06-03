-- reverse: modify "persistent_volumes" table
ALTER TABLE "public"."persistent_volumes" DROP COLUMN "cifs_config_gid", DROP COLUMN "cifs_config_uid";
