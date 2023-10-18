// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/ninedraft/bibliotheca/storage/ent/author"
)

// Author is the model entity for the Author schema.
type Author struct {
	config `json:"-"`
	// ID of the ent.
	ID int64 `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Bio holds the value of the "bio" field.
	Bio string `json:"bio,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the AuthorQuery when eager-loading is set.
	Edges        AuthorEdges `json:"edges"`
	selectValues sql.SelectValues
}

// AuthorEdges holds the relations/edges for other nodes in the graph.
type AuthorEdges struct {
	// Books holds the value of the books edge.
	Books []*Book `json:"books,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// BooksOrErr returns the Books value or an error if the edge
// was not loaded in eager-loading.
func (e AuthorEdges) BooksOrErr() ([]*Book, error) {
	if e.loadedTypes[0] {
		return e.Books, nil
	}
	return nil, &NotLoadedError{edge: "books"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Author) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case author.FieldID:
			values[i] = new(sql.NullInt64)
		case author.FieldName, author.FieldBio:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Author fields.
func (a *Author) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case author.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			a.ID = int64(value.Int64)
		case author.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				a.Name = value.String
			}
		case author.FieldBio:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field bio", values[i])
			} else if value.Valid {
				a.Bio = value.String
			}
		default:
			a.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Author.
// This includes values selected through modifiers, order, etc.
func (a *Author) Value(name string) (ent.Value, error) {
	return a.selectValues.Get(name)
}

// QueryBooks queries the "books" edge of the Author entity.
func (a *Author) QueryBooks() *BookQuery {
	return NewAuthorClient(a.config).QueryBooks(a)
}

// Update returns a builder for updating this Author.
// Note that you need to call Author.Unwrap() before calling this method if this Author
// was returned from a transaction, and the transaction was committed or rolled back.
func (a *Author) Update() *AuthorUpdateOne {
	return NewAuthorClient(a.config).UpdateOne(a)
}

// Unwrap unwraps the Author entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (a *Author) Unwrap() *Author {
	_tx, ok := a.config.driver.(*txDriver)
	if !ok {
		panic("ent: Author is not a transactional entity")
	}
	a.config.driver = _tx.drv
	return a
}

// String implements the fmt.Stringer.
func (a *Author) String() string {
	var builder strings.Builder
	builder.WriteString("Author(")
	builder.WriteString(fmt.Sprintf("id=%v, ", a.ID))
	builder.WriteString("name=")
	builder.WriteString(a.Name)
	builder.WriteString(", ")
	builder.WriteString("bio=")
	builder.WriteString(a.Bio)
	builder.WriteByte(')')
	return builder.String()
}

// Authors is a parsable slice of Author.
type Authors []*Author
