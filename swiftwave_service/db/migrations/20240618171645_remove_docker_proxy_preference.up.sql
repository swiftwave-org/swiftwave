-- modify "applications" table
ALTER TABLE "public"."applications" DROP COLUMN "docker_proxy_server_preference", DROP COLUMN "docker_proxy_specific_server_id", ADD COLUMN "preferred_server_hostnames" character varying(254) NULL;
