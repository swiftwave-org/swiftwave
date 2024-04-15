-- reverse: modify "persistent_volumes" table
ALTER TABLE "public"."persistent_volumes" DROP COLUMN "cifs_config_dir_mode", DROP COLUMN "cifs_config_file_mode", DROP COLUMN "cifs_config_password", DROP COLUMN "cifs_config_username", DROP COLUMN "cifs_config_host", DROP COLUMN "cifs_config_share";
