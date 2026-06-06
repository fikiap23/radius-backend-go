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

type Project struct {
	ent.Schema
}

func (Project) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "projects"},
		entsql.Checks(map[string]string{
			"projects_progress_range": "progress >= 0 AND progress <= 100",
		}),
	}
}

func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("workspace_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.String("icon").
			Default("🚀").
			NotEmpty(),
		field.Enum("cover").
			Values("emerald", "ocean", "sunset", "violet", "rose", "slate").
			Default("emerald"),
		field.String("cover_image_url").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.Enum("status").
			Values("active", "on_hold", "completed").
			Default("active"),
		field.Bool("is_favorite").
			Default(false),
		field.Time("archived_at").
			Optional().
			Nillable(),
		field.Int("open_tasks").
			Default(0).
			NonNegative(),
		field.Int("progress").
			Default(0).
			Min(0).
			Max(100),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workspace", Workspace.Type).
			Ref("projects").
			Field("workspace_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workspace_id"),
	}
}
