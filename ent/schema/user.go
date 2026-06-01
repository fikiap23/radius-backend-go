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

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
		entsql.Checks(map[string]string{
			"users_email_format": "email ~* '^[^@]+@[^@]+\\.[^@]+$'",
		}),
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{dialect.Postgres: "uuid"}).
			Default("gen_random_uuid()").
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("email").
			SchemaType(map[string]string{dialect.Postgres: "citext"}).
			NotEmpty().
			Unique(),
		field.String("password_hash").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.Time("email_verified_at").
			Optional().
			Nillable(),
		field.String("avatar_url").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.Time("last_login_at").
			Optional().
			Nillable(),
		field.String("timezone").
			SchemaType(map[string]string{dialect.Postgres: "varchar(64)"}).
			Optional().
			Nillable(),
		field.String("locale").
			SchemaType(map[string]string{dialect.Postgres: "varchar(10)"}).
			Default("en").
			NotEmpty(),
		field.Enum("status").
			Values("active", "inactive").
			Default("active"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("oauth_accounts", UserOAuthAccount.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
