package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// History holds the schema definition for the History entity.
type History struct {
	ent.Schema
}

// Fields of the History.
func (History) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive(),
		field.Int("amount").
			Default(0),
		field.Int("user_id").
			Optional(),
	}
}

// Edges of the History.
func (History) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("histories").
			Field("user_id").
			Unique(),
	}
}
