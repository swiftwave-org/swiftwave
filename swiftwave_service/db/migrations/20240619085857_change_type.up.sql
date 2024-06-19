-- modify "applications" table
ALTER TABLE "public"."applications" ALTER COLUMN "preferred_server_hostnames" TYPE text[];
