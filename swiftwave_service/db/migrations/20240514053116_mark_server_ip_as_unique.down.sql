-- reverse: modify "servers" table
ALTER TABLE "public"."servers" DROP CONSTRAINT "uni_servers_ip", ADD CONSTRAINT "uni_servers_host_name" UNIQUE ("host_name");
