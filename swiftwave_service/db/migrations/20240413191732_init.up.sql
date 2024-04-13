-- create "enqueued_tasks" table
CREATE TABLE "public"."enqueued_tasks" (
  "id" bigserial NOT NULL,
  "queue_name" text NULL,
  "body" text NULL,
  "hash" text NULL,
  PRIMARY KEY ("id")
);
-- create "system_configs" table
CREATE TABLE "public"."system_configs" (
  "id" bigserial NOT NULL,
  "config_version" bigint NULL DEFAULT 1,
  "network_name" text NULL,
  "restricted_ports" integer[] NULL,
  "jwt_secret_key" text NULL,
  "ssh_private_key" text NULL,
  "lets_encrypt_config_id" bigserial NOT NULL,
  "lets_encrypt_config_staging" boolean NULL DEFAULT false,
  "lets_encrypt_config_email_id" text NULL,
  "lets_encrypt_config_private_key" text NULL,
  "haproxy_config_image" text NULL,
  "haproxy_config_username" text NULL,
  "haproxy_config_password" text NULL,
  "udp_proxy_config_image" text NULL,
  "persistent_volume_backup_config_s3_backup_enabled" boolean NULL,
  "persistent_volume_backup_config_s3_backup_endpoint" text NULL,
  "persistent_volume_backup_config_s3_backup_region" text NULL,
  "persistent_volume_backup_config_s3_backup_bucket" text NULL,
  "persistent_volume_backup_config_s3_backup_access_key_id" text NULL,
  "persistent_volume_backup_config_s3_backup_secret_access_key" text NULL,
  "persistent_volume_backup_config_s3_backup_force_path_style" boolean NULL,
  "pub_sub_config_mode" text NULL DEFAULT 'local',
  "pub_sub_config_buffer_length" bigint NULL DEFAULT 2000,
  "pub_sub_config_redis_host" text NULL,
  "pub_sub_config_redis_port" bigint NULL,
  "pub_sub_config_redis_password" text NULL,
  "pub_sub_config_redis_database_id" bigint NULL,
  "task_queue_config_mode" text NULL DEFAULT 'local',
  "task_queue_config_remote_task_queue_type" text NULL DEFAULT 'none',
  "task_queue_config_max_outstanding_messages_per_queue" bigint NULL DEFAULT 2,
  "task_queue_config_no_of_workers_per_queue" bigint NULL,
  "task_queue_config_amqp_protocol" text NULL,
  "task_queue_config_amqp_host" text NULL,
  "task_queue_config_amqp_port" bigint NULL,
  "task_queue_config_amqp_user" text NULL,
  "task_queue_config_amqp_password" text NULL,
  "task_queue_config_amqp_v_host" text NULL,
  "task_queue_config_redis_host" text NULL,
  "task_queue_config_redis_port" bigint NULL,
  "task_queue_config_redis_password" text NULL,
  "task_queue_config_redis_database_id" bigint NULL,
  "image_registry_config_endpoint" text NULL,
  "image_registry_config_username" text NULL,
  "image_registry_config_password" text NULL,
  "image_registry_config_namespace" text NULL,
  PRIMARY KEY ("id", "lets_encrypt_config_id")
);
-- create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "username" text NULL,
  "role" text NULL DEFAULT 'user',
  "password_hash" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_users_username" UNIQUE ("username")
);
-- create "key_authorization_tokens" table
CREATE TABLE "public"."key_authorization_tokens" (
  "token" text NOT NULL,
  "authorization_token" text NULL,
  PRIMARY KEY ("token")
);
-- create "servers" table
CREATE TABLE "public"."servers" (
  "id" bigserial NOT NULL,
  "ip" text NULL,
  "host_name" text NULL,
  "user" text NULL,
  "ssh_port" bigint NULL DEFAULT 22,
  "schedule_deployments" boolean NULL DEFAULT true,
  "docker_unix_socket_path" text NULL,
  "swarm_mode" text NULL,
  "proxy_enabled" boolean NULL DEFAULT false,
  "proxy_setup_running" boolean NULL DEFAULT false,
  "proxy_type" text NULL DEFAULT 'active',
  "status" text NULL,
  "last_ping" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_servers_host_name" UNIQUE ("host_name")
);
-- create "analytics_service_tokens" table
CREATE TABLE "public"."analytics_service_tokens" (
  "id" text NOT NULL,
  "token" text NULL,
  "server_id" bigint NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_analytics_service_tokens_token" UNIQUE ("token"),
  CONSTRAINT "fk_servers_analytics_service_token" FOREIGN KEY ("server_id") REFERENCES "public"."servers" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "applications" table
CREATE TABLE "public"."applications" (
  "id" text NOT NULL,
  "name" text NULL,
  "deployment_mode" text NULL,
  "replicas" bigint NULL,
  "command" text NULL,
  "capabilities" text[] NULL,
  "sysctls" text[] NULL,
  "is_deleted" boolean NULL DEFAULT false,
  "webhook_token" text NULL,
  "is_sleeping" boolean NULL DEFAULT false,
  "application_group" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_applications_name" UNIQUE ("name")
);
-- create "application_service_resource_stats" table
CREATE TABLE "public"."application_service_resource_stats" (
  "id" bigserial NOT NULL,
  "application_id" text NULL,
  "cpu_usage_percent" smallint NULL,
  "reporting_server_count" bigint NULL,
  "used_memory_mb" bigint NULL,
  "network_sent_kb" bigint NULL,
  "network_recv_kb" bigint NULL,
  "network_sent_kbps" bigint NULL,
  "network_recv_kbps" bigint NULL,
  "recorded_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_applications_resource_stats" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "git_credentials" table
CREATE TABLE "public"."git_credentials" (
  "id" bigserial NOT NULL,
  "name" text NULL,
  "username" text NULL,
  "password" text NULL,
  PRIMARY KEY ("id")
);
-- create "image_registry_credentials" table
CREATE TABLE "public"."image_registry_credentials" (
  "id" bigserial NOT NULL,
  "url" text NULL,
  "username" text NULL,
  "password" text NULL,
  PRIMARY KEY ("id")
);
-- create "deployments" table
CREATE TABLE "public"."deployments" (
  "id" text NOT NULL,
  "application_id" text NULL,
  "upstream_type" text NULL,
  "git_credential_id" bigint NULL,
  "git_provider" text NULL,
  "repository_owner" text NULL,
  "repository_name" text NULL,
  "repository_branch" text NULL,
  "code_path" text NULL,
  "commit_hash" text NULL,
  "source_code_compressed_file_name" text NULL,
  "docker_image" text NULL,
  "image_registry_credential_id" bigint NULL,
  "dockerfile" text NULL,
  "status" text NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_applications_deployments" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_applications_latest_deployment" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_git_credentials_deployments" FOREIGN KEY ("git_credential_id") REFERENCES "public"."git_credentials" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_image_registry_credentials_deployments" FOREIGN KEY ("image_registry_credential_id") REFERENCES "public"."image_registry_credentials" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "build_args" table
CREATE TABLE "public"."build_args" (
  "id" bigserial NOT NULL,
  "deployment_id" text NULL,
  "key" text NULL,
  "value" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_deployments_build_args" FOREIGN KEY ("deployment_id") REFERENCES "public"."deployments" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "console_tokens" table
CREATE TABLE "public"."console_tokens" (
  "id" text NOT NULL,
  "target" text NULL,
  "server_id" bigint NULL,
  "application_id" text NULL,
  "token" text NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_console_tokens_token" UNIQUE ("token"),
  CONSTRAINT "fk_applications_console_tokens" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_servers_console_tokens" FOREIGN KEY ("server_id") REFERENCES "public"."servers" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "deployment_logs" table
CREATE TABLE "public"."deployment_logs" (
  "id" bigserial NOT NULL,
  "deployment_id" text NULL,
  "content" text NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_deployments_logs" FOREIGN KEY ("deployment_id") REFERENCES "public"."deployments" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "environment_variables" table
CREATE TABLE "public"."environment_variables" (
  "id" bigserial NOT NULL,
  "application_id" text NULL,
  "key" text NULL,
  "value" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_applications_environment_variables" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "domains" table
CREATE TABLE "public"."domains" (
  "id" bigserial NOT NULL,
  "name" text NULL,
  "ssl_status" text NULL,
  "ssl_private_key" text NULL,
  "ssl_full_chain" text NULL,
  "ssl_issued_at" timestamptz NULL,
  "ssl_expired_at" timestamptz NULL,
  "ssl_issuer" text NULL,
  "s_slauto_renew" boolean NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_domains_name" UNIQUE ("name")
);
-- create "ingress_rules" table
CREATE TABLE "public"."ingress_rules" (
  "id" bigserial NOT NULL,
  "domain_id" bigint NULL,
  "application_id" text NULL,
  "protocol" text NULL,
  "port" bigint NULL,
  "target_port" bigint NULL,
  "status" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_applications_ingress_rules" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_domains_ingress_rules" FOREIGN KEY ("domain_id") REFERENCES "public"."domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "persistent_volumes" table
CREATE TABLE "public"."persistent_volumes" (
  "id" bigserial NOT NULL,
  "name" text NULL,
  "type" text NULL DEFAULT 'local',
  "nfs_config_host" text NULL,
  "nfs_config_path" text NULL,
  "nfs_config_version" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_persistent_volumes_name" UNIQUE ("name")
);
-- create "persistent_volume_backups" table
CREATE TABLE "public"."persistent_volume_backups" (
  "id" bigserial NOT NULL,
  "type" text NULL,
  "status" text NULL,
  "file" text NULL,
  "file_size_mb" numeric NULL,
  "persistent_volume_id" bigint NULL,
  "created_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_persistent_volumes_persistent_volume_backups" FOREIGN KEY ("persistent_volume_id") REFERENCES "public"."persistent_volumes" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "persistent_volume_bindings" table
CREATE TABLE "public"."persistent_volume_bindings" (
  "id" bigserial NOT NULL,
  "application_id" text NULL,
  "persistent_volume_id" bigint NULL,
  "mounting_path" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_applications_persistent_volume_bindings" FOREIGN KEY ("application_id") REFERENCES "public"."applications" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_persistent_volumes_persistent_volume_bindings" FOREIGN KEY ("persistent_volume_id") REFERENCES "public"."persistent_volumes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "persistent_volume_restores" table
CREATE TABLE "public"."persistent_volume_restores" (
  "id" bigserial NOT NULL,
  "type" text NULL,
  "file" text NULL,
  "status" text NULL,
  "persistent_volume_id" bigint NULL,
  "created_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_persistent_volumes_persistent_volume_restores" FOREIGN KEY ("persistent_volume_id") REFERENCES "public"."persistent_volumes" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create "redirect_rules" table
CREATE TABLE "public"."redirect_rules" (
  "id" bigserial NOT NULL,
  "domain_id" bigint NULL,
  "protocol" text NULL,
  "redirect_url" text NULL,
  "status" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_domains_redirect_rules" FOREIGN KEY ("domain_id") REFERENCES "public"."domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create "server_logs" table
CREATE TABLE "public"."server_logs" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "server_id" bigint NULL,
  "title" text NULL,
  "content" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_servers_logs" FOREIGN KEY ("server_id") REFERENCES "public"."servers" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- create index "idx_server_logs_deleted_at" to table: "server_logs"
CREATE INDEX "idx_server_logs_deleted_at" ON "public"."server_logs" ("deleted_at");
-- create "server_resource_stats" table
CREATE TABLE "public"."server_resource_stats" (
  "id" bigserial NOT NULL,
  "server_id" bigint NULL,
  "cpu_usage_percent" smallint NULL,
  "memory_total_gb" numeric NULL,
  "memory_used_gb" numeric NULL,
  "memory_cached_gb" numeric NULL,
  "disk_stats" bytea NULL,
  "network_sent_kb" bigint NULL,
  "network_recv_kb" bigint NULL,
  "network_sent_kbps" bigint NULL,
  "network_recv_kbps" bigint NULL,
  "recorded_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_servers_resource_stats" FOREIGN KEY ("server_id") REFERENCES "public"."servers" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
