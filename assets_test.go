package main

import (
	"strings"
	"testing"
	"time"

	"github.com/ninedraft/bibliotheca/storage/ent"
)

func TestBooks(t *testing.T) {
	t.Parallel()

	got := &strings.Builder{}

	err := assets.ExecuteTemplate(got, "books.html", &booksView{
		Books: []*ent.Book{{
			ID:        1,
			Title:     "The Lord of the Rings",
			WrittenAt: time.Date(1954, 7, 29, 0, 0, 0, 0, time.UTC).Unix(),
		}},
		Authors: map[int64][]*ent.Author{
			1: {
				{Name: "J. R. R. Tolkien"},
			},
		},
	})

	if err != nil {
		t.Fatalf("template: %s", err)
	}

	t.Log(got)
}
