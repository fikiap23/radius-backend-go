-- Create "tasks" table
CREATE TABLE "tasks" ("id" uuid NOT NULL, "title" character varying NOT NULL, "description" text NULL, "status" character varying NOT NULL DEFAULT 'todo', "priority" character varying NOT NULL DEFAULT 'medium', "due_at" timestamptz NULL, "label_ids" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "column_id" uuid NULL, "project_id" uuid NOT NULL, "assignee_id" uuid NULL, "workspace_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "tasks_board_columns_tasks" FOREIGN KEY ("column_id") REFERENCES "board_columns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "tasks_projects_tasks" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "tasks_users_assigned_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "tasks_workspaces_tasks" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "task_assignee_id" to table: "tasks"
CREATE INDEX "task_assignee_id" ON "tasks" ("assignee_id");
-- Create index "task_column_id" to table: "tasks"
CREATE INDEX "task_column_id" ON "tasks" ("column_id");
-- Create index "task_project_id" to table: "tasks"
CREATE INDEX "task_project_id" ON "tasks" ("project_id");
-- Create index "task_workspace_id" to table: "tasks"
CREATE INDEX "task_workspace_id" ON "tasks" ("workspace_id");
-- Create "task_activity_logs" table
CREATE TABLE "task_activity_logs" ("id" uuid NOT NULL, "title" character varying NOT NULL, "description" character varying NULL, "icon" character varying NOT NULL, "occurred_at" timestamptz NOT NULL, "task_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "task_activity_logs_tasks_activity_logs" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "taskactivitylog_occurred_at" to table: "task_activity_logs"
CREATE INDEX "taskactivitylog_occurred_at" ON "task_activity_logs" ("occurred_at");
-- Create index "taskactivitylog_task_id" to table: "task_activity_logs"
CREATE INDEX "taskactivitylog_task_id" ON "task_activity_logs" ("task_id");
-- Create "task_attachments" table
CREATE TABLE "task_attachments" ("id" uuid NOT NULL, "name" character varying NOT NULL, "size" bigint NOT NULL, "mime_type" character varying(120) NOT NULL, "storage_key" text NOT NULL, "url" text NOT NULL, "uploaded_at" timestamptz NOT NULL, "task_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "task_attachments_tasks_attachments" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "taskattachment_task_id" to table: "task_attachments"
CREATE INDEX "taskattachment_task_id" ON "task_attachments" ("task_id");
-- Create "task_checklist_items" table
CREATE TABLE "task_checklist_items" ("id" uuid NOT NULL, "text" character varying NOT NULL, "checked" boolean NOT NULL DEFAULT false, "task_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "task_checklist_items_tasks_checklist_items" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "taskchecklistitem_task_id" to table: "task_checklist_items"
CREATE INDEX "taskchecklistitem_task_id" ON "task_checklist_items" ("task_id");
-- Create "task_comments" table
CREATE TABLE "task_comments" ("id" uuid NOT NULL, "author_name" character varying NOT NULL, "body" text NOT NULL, "mention_ids" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "task_id" uuid NOT NULL, "author_id" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "task_comments_tasks_comments" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "task_comments_users_task_comments" FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "taskcomment_author_id" to table: "task_comments"
CREATE INDEX "taskcomment_author_id" ON "task_comments" ("author_id");
-- Create index "taskcomment_task_id" to table: "task_comments"
CREATE INDEX "taskcomment_task_id" ON "task_comments" ("task_id");
-- Create "task_subtasks" table
CREATE TABLE "task_subtasks" ("id" uuid NOT NULL, "title" character varying NOT NULL, "done" boolean NOT NULL DEFAULT false, "task_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "task_subtasks_tasks_subtasks" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "tasksubtask_task_id" to table: "task_subtasks"
CREATE INDEX "tasksubtask_task_id" ON "task_subtasks" ("task_id");
