-- modify "applications" table
ALTER TABLE "public"."applications" ADD COLUMN "resource_limit_memory_mb" bigint NULL DEFAULT 0, ADD COLUMN "reserved_resource_memory_mb" bigint NULL DEFAULT 0;
