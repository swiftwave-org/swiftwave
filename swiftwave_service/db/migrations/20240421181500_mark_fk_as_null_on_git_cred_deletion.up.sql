-- modify "deployments" table
ALTER TABLE "public"."deployments" DROP CONSTRAINT "fk_git_credentials_deployments", ADD
 CONSTRAINT "fk_git_credentials_deployments" FOREIGN KEY ("git_credential_id") REFERENCES "public"."git_credentials" ("id") ON UPDATE CASCADE ON DELETE SET NULL;
