// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// AuthorsColumns holds the columns for the "authors" table.
	AuthorsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt64, Increment: true},
		{Name: "name", Type: field.TypeString},
		{Name: "bio", Type: field.TypeString, Nullable: true},
	}
	// AuthorsTable holds the schema information for the "authors" table.
	AuthorsTable = &schema.Table{
		Name:       "authors",
		Columns:    AuthorsColumns,
		PrimaryKey: []*schema.Column{AuthorsColumns[0]},
	}
	// BooksColumns holds the columns for the "books" table.
	BooksColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt64, Increment: true},
		{Name: "title", Type: field.TypeString},
		{Name: "written_at", Type: field.TypeInt64},
		{Name: "cover_id", Type: field.TypeString, Nullable: true},
		{Name: "file_id", Type: field.TypeString, Nullable: true},
	}
	// BooksTable holds the schema information for the "books" table.
	BooksTable = &schema.Table{
		Name:       "books",
		Columns:    BooksColumns,
		PrimaryKey: []*schema.Column{BooksColumns[0]},
	}
	// BookAuthorsColumns holds the columns for the "book_authors" table.
	BookAuthorsColumns = []*schema.Column{
		{Name: "book_id", Type: field.TypeInt64},
		{Name: "author_id", Type: field.TypeInt64},
	}
	// BookAuthorsTable holds the schema information for the "book_authors" table.
	BookAuthorsTable = &schema.Table{
		Name:       "book_authors",
		Columns:    BookAuthorsColumns,
		PrimaryKey: []*schema.Column{BookAuthorsColumns[0], BookAuthorsColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "book_authors_book_id",
				Columns:    []*schema.Column{BookAuthorsColumns[0]},
				RefColumns: []*schema.Column{BooksColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "book_authors_author_id",
				Columns:    []*schema.Column{BookAuthorsColumns[1]},
				RefColumns: []*schema.Column{AuthorsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		AuthorsTable,
		BooksTable,
		BookAuthorsTable,
	}
)

func init() {
	BookAuthorsTable.ForeignKeys[0].RefTable = BooksTable
	BookAuthorsTable.ForeignKeys[1].RefTable = AuthorsTable
}
