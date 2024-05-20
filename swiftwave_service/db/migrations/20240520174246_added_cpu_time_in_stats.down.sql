-- reverse: modify "application_service_resource_stats" table
ALTER TABLE "public"."application_service_resource_stats" DROP COLUMN "system_cpu_time", DROP COLUMN "service_cpu_time";
