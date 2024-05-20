-- modify "application_service_resource_stats" table
ALTER TABLE "public"."application_service_resource_stats" ADD COLUMN "service_cpu_time" bigint NULL, ADD COLUMN "system_cpu_time" bigint NULL;
