package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.Int("stock").Positive().Default(0),
		field.String("title").Default("unknown"),
		field.Enum("sale_status").
			Values("0", "1").Default("1"),
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return nil
}
