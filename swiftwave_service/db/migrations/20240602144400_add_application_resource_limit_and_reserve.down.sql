-- reverse: modify "applications" table
ALTER TABLE "public"."applications" DROP COLUMN "reserved_resource_memory_mb", DROP COLUMN "resource_limit_memory_mb";
