-- modify "servers" table
ALTER TABLE "public"."servers" DROP CONSTRAINT "uni_servers_host_name", ADD CONSTRAINT "uni_servers_ip" UNIQUE ("ip");
