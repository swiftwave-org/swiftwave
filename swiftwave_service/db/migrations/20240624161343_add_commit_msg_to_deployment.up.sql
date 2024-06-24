-- modify "deployments" table
ALTER TABLE "public"."deployments" ADD COLUMN "commit_message" text NULL;
