-- Create "board_columns" table
CREATE TABLE "board_columns" ("id" uuid NOT NULL, "title" character varying NOT NULL, "status" character varying(64) NOT NULL, "wip_limit" bigint NULL, "sort_order" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "project_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "board_columns_projects_board_columns" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "boardcolumn_project_id" to table: "board_columns"
CREATE INDEX "boardcolumn_project_id" ON "board_columns" ("project_id");
-- Create index "boardcolumn_project_id_sort_order" to table: "board_columns"
CREATE INDEX "boardcolumn_project_id_sort_order" ON "board_columns" ("project_id", "sort_order");
-- Create index "boardcolumn_project_id_status" to table: "board_columns"
CREATE UNIQUE INDEX "boardcolumn_project_id_status" ON "board_columns" ("project_id", "status");
