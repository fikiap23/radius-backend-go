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

type TaskActivityLog struct {
	ent.Schema
}

func (TaskActivityLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "task_activity_logs"},
	}
}

func (TaskActivityLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("task_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("title").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.String("icon").
			NotEmpty(),
		field.Time("occurred_at").
			Default(time.Now).
			Immutable(),
	}
}

func (TaskActivityLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).
			Ref("activity_logs").
			Field("task_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (TaskActivityLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("occurred_at"),
	}
}
