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

type WorkspaceMember struct {
	ent.Schema
}

func (WorkspaceMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workspace_members"},
	}
}

func (WorkspaceMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("workspace_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("user_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Optional().
			Nillable(),
		field.String("name").
			NotEmpty(),
		field.String("email").
			SchemaType(map[string]string{dialect.Postgres: "citext"}).
			NotEmpty(),
		field.Enum("role").
			Values("owner", "admin", "member", "viewer"),
		field.Enum("status").
			Values("active", "pending"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (WorkspaceMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).
			Ref("members").
			Field("workspace_id").
			Required().
			Unique().
			Immutable(),
		edge.From("user", User.Type).
			Ref("workspace_members").
			Field("user_id").
			Unique(),
	}
}

func (WorkspaceMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id"),
		index.Fields("user_id"),
		index.Fields("workspace_id", "email").
			Unique(),
	}
}
