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

type BoardColumn struct {
	ent.Schema
}

func (BoardColumn) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "board_columns"},
	}
}

func (BoardColumn) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("project_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("title").
			NotEmpty(),
		field.String("status").
			SchemaType(map[string]string{dialect.Postgres: "varchar(64)"}).
			NotEmpty(),
		field.Int("wip_limit").
			Optional().
			Nillable().
			NonNegative(),
		field.Int("sort_order").
			Default(0).
			NonNegative(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (BoardColumn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("board_columns").
			Field("project_id").
			Required().
			Unique().
			Immutable(),
		edge.To("tasks", Task.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (BoardColumn) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("project_id", "status").
			Unique(),
		index.Fields("project_id", "sort_order"),
	}
}
