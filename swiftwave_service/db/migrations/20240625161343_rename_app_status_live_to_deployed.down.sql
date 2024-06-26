-- reverse: modify "deployments" table
UPDATE "public"."deployments" SET status = 'live' WHERE status = 'deployed';
