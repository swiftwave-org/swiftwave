-- modify "applications" table
ALTER TABLE "public"."applications" ADD COLUMN "preferred_server_hostnames" character varying(254) NULL, ADD COLUMN "docker_proxy_enabled" boolean NULL DEFAULT false, ADD COLUMN "docker_proxy_permission_ping" text NULL DEFAULT 'read', ADD COLUMN "docker_proxy_permission_version" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_info" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_events" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_auth" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_secrets" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_build" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_commit" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_configs" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_containers" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_distribution" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_exec" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_grpc" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_images" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_networks" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_nodes" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_plugins" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_services" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_session" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_swarm" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_system" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_tasks" text NULL DEFAULT 'none', ADD COLUMN "docker_proxy_permission_volumes" text NULL DEFAULT 'none';
