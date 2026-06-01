-- Modify "user_oauth_accounts" table
ALTER TABLE "user_oauth_accounts" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "created_at" DROP DEFAULT, ALTER COLUMN "updated_at" DROP DEFAULT;
-- Modify "users" table
ALTER TABLE "users" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "created_at" DROP DEFAULT, ALTER COLUMN "updated_at" DROP DEFAULT, ADD COLUMN "status" character varying NOT NULL DEFAULT 'active';
