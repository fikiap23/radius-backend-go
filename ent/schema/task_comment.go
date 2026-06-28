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

type TaskComment struct {
	ent.Schema
}

func (TaskComment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "task_comments"},
	}
}

func (TaskComment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("task_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("author_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Optional().
			Nillable(),
		field.String("author_name").
			NotEmpty(),
		field.String("body").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty(),
		field.JSON("mention_ids", []string{}).
			Default([]string{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (TaskComment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).
			Ref("comments").
			Field("task_id").
			Required().
			Unique().
			Immutable(),
		edge.From("author", User.Type).
			Ref("task_comments").
			Field("author_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (TaskComment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("author_id"),
	}
}
