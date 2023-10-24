package service

import (
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/go-chi/chi/v5"
	binding "github.com/gorilla/schema"
	"github.com/ninedraft/bibliotheca/storage/ent"
	"github.com/ninedraft/bibliotheca/storage/ent/author"
	"github.com/ninedraft/bibliotheca/storage/ent/book"
)

type Service struct {
	Storage *ent.Client
	Templ   *template.Template
	Static  fs.FS
}

func (srv *Service) BuildRoutes(mux chi.Router) {
	mux.Handle("/static/*", http.FileServer(http.FS(srv.Static)))
	mux.Route("/books", func(r chi.Router) {
		r.Get("/", srv.getBooks)
		r.Post("/", srv.createBook)
		r.Get("/new", srv.getBookForm)
	})

	mux.Route("/authors", func(r chi.Router) {
		r.Get("/", srv.listAuthors)
		r.Post("/", srv.createAuthor)
		r.Get("/new", srv.getAuthorForm)
	})
}

type booksView struct {
	Books   []*ent.Book
	Authors map[int64][]*ent.Author
}

func (view *booksView) List() []any {
	type bookView struct {
		Title     string
		WrittenAt time.Time
		Authors   []string
	}

	var list []any
	for _, book := range view.Books {
		authors := view.Authors[book.ID]
		var names []string
		for _, author := range authors {
			names = append(names, author.Name)
		}
		list = append(list, bookView{
			Title:     book.Title,
			WrittenAt: time.Unix(book.WrittenAt, 0),
			Authors:   names,
		})
	}

	return list
}

func (srv *Service) getBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query().Get("q")
	query := srv.Storage.Book.Query()

	if q != "" {
		query = query.Where(book.Or(
			book.TitleContainsFold(q),
			book.HasAuthorsWith(author.NameContainsFold(q)),
		))
	}

	books, err := query.WithAuthors().All(ctx)
	if err != nil {
		http.Error(w, "form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	bookAuthors := make(map[int64][]*ent.Author)
	for _, book := range books {
		authors := book.QueryAuthors().AllX(ctx)
		bookAuthors[book.ID] = authors
	}

	data := &booksView{
		Books:   books,
		Authors: bookAuthors,
	}

	if err := srv.Templ.ExecuteTemplate(w, "books.html", data); err != nil {
		log.Printf("ERROR: template: %v", err)
		return
	}
}

func (srv *Service) getBookForm(w http.ResponseWriter, r *http.Request) {
	authors, err := srv.Storage.Author.Query().All(r.Context())
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Authors": authors,
	}

	if r.URL.Query().Get("error") != "" {
		data["Error"] = r.URL.Query().Get("error")
	}

	if err := srv.Templ.ExecuteTemplate(w, "books_create.html", data); err != nil {
		log.Printf("ERROR: books_create.html: %s", err)
		return
	}
}

type bookForm struct {
	Title     string  `schema:"title"`
	WrittenAt Date    `schema:"written_at"`
	Authors   []int64 `schema:"authors"`
}

var binder = binding.NewDecoder()

type Date struct{ time.Time }

func init() {
	binder.IgnoreUnknownKeys(true)
	binder.RegisterConverter(Date{}, func(s string) reflect.Value {
		t, err := time.Parse(time.DateOnly, s)
		if err != nil {
			return reflect.Value{}
		}
		return reflect.ValueOf(Date{t})
	})
}

func (srv *Service) createBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form: "+err.Error(), http.StatusBadRequest)
		return
	}

	var book bookForm
	errBind := binder.Decode(&book, r.PostForm)
	if errBind != nil {
		withError(w, r, "/books/new", errBind)
		return
	}

	if book.WrittenAt.After(time.Now()) {
		withError(w, r, "/books/new", errors.New("book written in future"))
		return
	}

	bookCreation := srv.Storage.Book.Create().
		SetTitle(book.Title).
		SetWrittenAt(book.WrittenAt.Unix())

	bookCreation.Mutation().AddAuthorIDs(book.Authors...)

	_, err := bookCreation.Save(r.Context())
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	srv.getBooks(w, r)
}

func (srv *Service) getAuthorForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}

	if r.URL.Query().Get("error") != "" {
		data["Error"] = r.URL.Query().Get("error")
	}

	if err := srv.Templ.ExecuteTemplate(w, "authors_create.html", data); err != nil {
		log.Printf("ERROR: authors_create.html: %s", err)
		return
	}
}

type authorForm struct {
	Name string `schema:"name"`
	Bio  string `schema:"bio"`
}

func (srv *Service) createAuthor(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form: "+err.Error(), http.StatusBadRequest)
		return
	}

	var form authorForm
	errForm := binder.Decode(&form, r.PostForm)
	if errForm != nil {
		withError(w, r, "/authors/new", errForm)
		return
	}

	log.Println("creating author", form)

	_, errAuthor := srv.Storage.Author.Create().
		SetName(form.Name).
		SetBio(form.Bio).
		Save(r.Context())
	if errAuthor != nil {
		http.Error(w, "db: "+errAuthor.Error(), http.StatusInternalServerError)
		return
	}

	srv.listAuthors(w, r)
}

func (srv *Service) listAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authors, err := srv.Storage.Author.Query().All(ctx)
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := srv.Templ.ExecuteTemplate(w, "authors", authors); err != nil {
		log.Printf("ERROR: template: %v", err)
		return
	}
}

func withError(rw http.ResponseWriter, req *http.Request, to string, err error) {
	if err == nil {
		return
	}

	dst := &url.URL{
		Path: to,
		RawQuery: url.Values{
			"error": {err.Error()},
		}.Encode(),
	}

	http.Redirect(rw, req, dst.String(), http.StatusSeeOther)
}
