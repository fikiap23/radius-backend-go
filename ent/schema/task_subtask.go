package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TaskSubtask struct {
	ent.Schema
}

func (TaskSubtask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "task_subtasks"},
	}
}

func (TaskSubtask) Fields() []ent.Field {
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
		field.Bool("done").
			Default(false),
	}
}

func (TaskSubtask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).
			Ref("subtasks").
			Field("task_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (TaskSubtask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
	}
}
