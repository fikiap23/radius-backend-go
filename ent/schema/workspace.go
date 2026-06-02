package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Workspace struct {
	ent.Schema
}

func (Workspace) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workspaces"},
	}
}

func (Workspace) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("slug").
			SchemaType(map[string]string{dialect.Postgres: "varchar(255)"}).
			NotEmpty().
			Unique(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Workspace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", WorkspaceMember.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
