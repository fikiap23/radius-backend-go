-- Create "workspaces" table
CREATE TABLE "workspaces" ("id" uuid NOT NULL, "name" character varying NOT NULL, "slug" character varying(255) NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- Create index "workspaces_slug_key" to table: "workspaces"
CREATE UNIQUE INDEX "workspaces_slug_key" ON "workspaces" ("slug");
-- Create "workspace_members" table
CREATE TABLE "workspace_members" ("id" uuid NOT NULL, "name" character varying NOT NULL, "email" citext NOT NULL, "role" character varying NOT NULL, "status" character varying NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "user_id" uuid NULL, "workspace_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "workspace_members_users_workspace_members" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "workspace_members_workspaces_members" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "workspacemember_user_id" to table: "workspace_members"
CREATE INDEX "workspacemember_user_id" ON "workspace_members" ("user_id");
-- Create index "workspacemember_workspace_id" to table: "workspace_members"
CREATE INDEX "workspacemember_workspace_id" ON "workspace_members" ("workspace_id");
-- Create index "workspacemember_workspace_id_email" to table: "workspace_members"
CREATE UNIQUE INDEX "workspacemember_workspace_id_email" ON "workspace_members" ("workspace_id", "email");
