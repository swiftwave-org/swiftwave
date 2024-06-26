-- modify "deployments" table
UPDATE "public"."deployments" SET status = 'deployed' WHERE status = 'live';
