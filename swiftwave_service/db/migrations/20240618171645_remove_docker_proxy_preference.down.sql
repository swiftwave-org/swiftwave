-- reverse: modify "applications" table
ALTER TABLE "public"."applications" DROP COLUMN "preferred_server_hostnames", ADD COLUMN "docker_proxy_specific_server_id" bigint NULL, ADD COLUMN "docker_proxy_server_preference" text NULL DEFAULT 'any';
