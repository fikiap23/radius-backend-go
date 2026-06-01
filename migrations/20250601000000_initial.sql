-- Create the citext extension for case-insensitive email comparison.
CREATE EXTENSION IF NOT EXISTS citext;

-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying NOT NULL,
  "email" citext NOT NULL,
  "password_hash" text NULL,
  "email_verified_at" timestamptz NULL,
  "avatar_url" text NULL,
  "last_login_at" timestamptz NULL,
  "timezone" character varying(64) NULL,
  "locale" character varying(10) NOT NULL DEFAULT 'en',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email"),
  CONSTRAINT "users_email_format" CHECK (email ~* '^[^@]+@[^@]+\.[^@]+$')
);
CREATE INDEX "user_deleted_at" ON "users" ("deleted_at");

-- Create "user_oauth_accounts" table
CREATE TABLE "user_oauth_accounts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "provider" character varying(32) NOT NULL,
  "provider_user_id" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "user_oauth_accounts_users_oauth_accounts" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE INDEX "useroauthaccount_user_id" ON "user_oauth_accounts" ("user_id");
CREATE UNIQUE INDEX "useroauthaccount_provider_provider_user_id" ON "user_oauth_accounts" ("provider", "provider_user_id");
