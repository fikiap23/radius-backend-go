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

type TaskAttachment struct {
	ent.Schema
}

func (TaskAttachment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "task_attachments"},
	}
}

func (TaskAttachment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("task_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.Int64("size").
			NonNegative(),
		field.String("mime_type").
			SchemaType(map[string]string{dialect.Postgres: "varchar(120)"}).
			NotEmpty(),
		field.String("storage_key").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty(),
		field.String("url").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty(),
		field.Time("uploaded_at").
			Default(time.Now).
			Immutable(),
	}
}

func (TaskAttachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).
			Ref("attachments").
			Field("task_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (TaskAttachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
	}
}
