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

type UserOAuthAccount struct {
	ent.Schema
}

func (UserOAuthAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_oauth_accounts"},
	}
}

func (UserOAuthAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("user_id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Immutable(),
		field.String("provider").
			SchemaType(map[string]string{dialect.Postgres: "varchar(32)"}).
			NotEmpty(),
		field.String("provider_user_id").
			SchemaType(map[string]string{dialect.Postgres: "varchar(255)"}).
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (UserOAuthAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("oauth_accounts").
			Field("user_id").
			Required().
			Unique().
			Immutable(),
	}
}

func (UserOAuthAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("provider", "provider_user_id").
			Unique(),
	}
}
