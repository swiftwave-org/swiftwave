-- modify "persistent_volumes" table
ALTER TABLE "public"."persistent_volumes" ADD COLUMN "cifs_config_share" text NULL, ADD COLUMN "cifs_config_host" text NULL, ADD COLUMN "cifs_config_username" text NULL, ADD COLUMN "cifs_config_password" text NULL, ADD COLUMN "cifs_config_file_mode" text NULL, ADD COLUMN "cifs_config_dir_mode" text NULL;
