-- reverse: rename a column from "s_slauto_renew" to "ssl_auto_renew"
ALTER TABLE "public"."domains" RENAME COLUMN "ssl_auto_renew" TO "s_slauto_renew";
