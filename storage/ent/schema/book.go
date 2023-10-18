package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Book holds the schema definition for the Book entity.
type Book struct {
	ent.Schema
}

// Fields of the Book.
func (Book) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique(),
		field.String("title"),
		field.Int64("written_at").DefaultFunc(now),
		field.String("cover_id").Optional(),
		field.String("file_id").Optional(),
	}
}

// Edges of the Book.
func (Book) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("authors", Author.Type),
	}
}

func now() int64 {
	return time.Now().Unix()
}
