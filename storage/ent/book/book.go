// Code generated by ent, DO NOT EDIT.

package book

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the book type in the database.
	Label = "book"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldWrittenAt holds the string denoting the written_at field in the database.
	FieldWrittenAt = "written_at"
	// FieldCoverID holds the string denoting the cover_id field in the database.
	FieldCoverID = "cover_id"
	// FieldFileID holds the string denoting the file_id field in the database.
	FieldFileID = "file_id"
	// EdgeAuthors holds the string denoting the authors edge name in mutations.
	EdgeAuthors = "authors"
	// Table holds the table name of the book in the database.
	Table = "books"
	// AuthorsTable is the table that holds the authors relation/edge. The primary key declared below.
	AuthorsTable = "book_authors"
	// AuthorsInverseTable is the table name for the Author entity.
	// It exists in this package in order to avoid circular dependency with the "author" package.
	AuthorsInverseTable = "authors"
)

// Columns holds all SQL columns for book fields.
var Columns = []string{
	FieldID,
	FieldTitle,
	FieldWrittenAt,
	FieldCoverID,
	FieldFileID,
}

var (
	// AuthorsPrimaryKey and AuthorsColumn2 are the table columns denoting the
	// primary key for the authors relation (M2M).
	AuthorsPrimaryKey = []string{"book_id", "author_id"}
)

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultWrittenAt holds the default value on creation for the "written_at" field.
	DefaultWrittenAt func() int64
)

// OrderOption defines the ordering options for the Book queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByTitle orders the results by the title field.
func ByTitle(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTitle, opts...).ToFunc()
}

// ByWrittenAt orders the results by the written_at field.
func ByWrittenAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldWrittenAt, opts...).ToFunc()
}

// ByCoverID orders the results by the cover_id field.
func ByCoverID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCoverID, opts...).ToFunc()
}

// ByFileID orders the results by the file_id field.
func ByFileID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFileID, opts...).ToFunc()
}

// ByAuthorsCount orders the results by authors count.
func ByAuthorsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAuthorsStep(), opts...)
	}
}

// ByAuthors orders the results by authors terms.
func ByAuthors(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAuthorsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newAuthorsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AuthorsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, AuthorsTable, AuthorsPrimaryKey...),
	)
}
