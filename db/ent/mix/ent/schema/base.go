package schema

import (
	"miopkg/util/snowflake"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Base holds the schema definition for the Base entity.
type Base struct {
	// ent.Schema
	mixin.Schema
	// Schema
}

// Fields of the Base.
func (Base) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("mid").
			Default(snowflake.GenInt64()).
			Unique(),
		field.Bool("deleted"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).SchemaType(map[string]string{
			dialect.MySQL: "datetime",
		}),
		field.Time("updated_at").
			Default(time.Now).SchemaType(map[string]string{
			dialect.MySQL: "datetime",
		}).
			UpdateDefault(time.Now),
	}
}

// Edges of the Base.
// func (Base) Edges() []ent.Edge {
// 	return nil
// }
// func (Base) Indexes() []ent.Index {
// 	return []ent.Index{
// 		index.Fields("mid").
// 			Unique(),
// 	}
// }
