-- rename a column from "s_slauto_renew" to "ssl_auto_renew"
ALTER TABLE "public"."domains" RENAME COLUMN "s_slauto_renew" TO "ssl_auto_renew";
