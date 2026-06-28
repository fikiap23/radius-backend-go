package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Task struct {
	ent.Schema
}

func (Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tasks"},
	}
}

func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("project_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("workspace_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("title").
			NotEmpty(),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.Enum("status").
			Values("backlog", "todo", "in_progress", "review", "done").
			Default("todo"),
		field.String("column_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Optional().
			Nillable(),
		field.Enum("priority").
			Values("low", "medium", "high", "urgent").
			Default("medium"),
		field.Time("due_at").
			Optional().
			Nillable(),
		field.JSON("label_ids", []string{}).
			Default([]string{}),
		field.String("assignee_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("tasks").
			Field("project_id").
			Required().
			Unique().
			Immutable(),
		edge.From("workspace", Workspace.Type).
			Ref("tasks").
			Field("workspace_id").
			Required().
			Unique().
			Immutable(),
		edge.From("column", BoardColumn.Type).
			Ref("tasks").
			Field("column_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.From("assignee", User.Type).
			Ref("assigned_tasks").
			Field("assignee_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("subtasks", TaskSubtask.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("checklist_items", TaskChecklistItem.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("attachments", TaskAttachment.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("comments", TaskComment.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("activity_logs", TaskActivityLog.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Task) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("workspace_id"),
		index.Fields("column_id"),
		index.Fields("assignee_id"),
	}
}
