-- reverse: modify "applications" table
ALTER TABLE "public"."applications" ADD COLUMN "docker_proxy_authentication_token" text NULL;
