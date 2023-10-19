package main

import (
	"context"
	dbsql "database/sql"
	"embed"
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/go-chi/chi/v5"
	binding "github.com/gorilla/schema"
	"github.com/ninedraft/bibliotheca/storage/ent"
	"github.com/ninedraft/bibliotheca/storage/ent/author"
	"github.com/ninedraft/bibliotheca/storage/ent/book"

	_ "modernc.org/sqlite"
)

func main() {
	addr := "localhost:8080"
	flag.StringVar(&addr, "addr", addr, "server address")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	go func() {
		<-ctx.Done()
		time.Sleep(5 * time.Second)
		log.Println("force shutdown")
		os.Exit(1)
	}()

	db, errOpenDB := dbsql.Open("sqlite",
		"file:bib.sqlite?cache=shared&_pragma=foreign_keys(1)")
	if errOpenDB != nil {
		panic("db: " + errOpenDB.Error())
	}
	defer func() { _ = db.Close() }()

	drv := sql.OpenDB(dialect.SQLite, db)
	defer func() { _ = drv.Close() }()

	client := ent.NewClient(
		ent.Driver(drv),
		ent.Log(log.Println),
	)
	defer func() { _ = client.Close() }()

	if errMigrate := client.Schema.Create(ctx); errMigrate != nil {
		panic("migrate: " + errMigrate.Error())
	}

	srv := &service{
		storage: client,
	}
	mux := chi.NewMux().With(logMW)

	srv.routes(mux)

	log.Printf("starting server at %s", addr)
	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Addr:              addr,
		Handler:           mux,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down service")
		_ = server.Shutdown(context.Background())
	}()

	errServe := server.ListenAndServe()
	if errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
		panic("server: " + errServe.Error())
	}
}

var (
	//go:embed static/*
	static embed.FS

	//go:embed templ/*
	assetsFS embed.FS
	assets   = template.Must(template.ParseFS(assetsFS, "templ/*.html"))
)

type service struct {
	storage *ent.Client
}

func (srv *service) routes(mux chi.Router) {
	mux.Handle("/static/*", http.FileServer(http.FS(static)))
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

func (srv *service) getBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query().Get("q")
	query := srv.storage.Book.Query()

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

	if err := assets.ExecuteTemplate(w, "books.html", data); err != nil {
		log.Printf("ERROR: template: %v", err)
		return
	}
}

func (srv *service) getBookForm(w http.ResponseWriter, r *http.Request) {
	authors, err := srv.storage.Author.Query().All(r.Context())
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

	if err := assets.ExecuteTemplate(w, "books_create.html", data); err != nil {
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

func (srv *service) createBook(w http.ResponseWriter, r *http.Request) {
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

	bookCreation := srv.storage.Book.Create().
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

func (srv *service) getAuthorForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}

	if r.URL.Query().Get("error") != "" {
		data["Error"] = r.URL.Query().Get("error")
	}

	if err := assets.ExecuteTemplate(w, "authors_create.html", data); err != nil {
		log.Printf("ERROR: authors_create.html: %s", err)
		return
	}
}

type authorForm struct {
	Name string `schema:"name"`
	Bio  string `schema:"bio"`
}

func (srv *service) createAuthor(w http.ResponseWriter, r *http.Request) {
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

	_, errAuthor := srv.storage.Author.Create().
		SetName(form.Name).
		SetBio(form.Bio).
		Save(r.Context())
	if errAuthor != nil {
		http.Error(w, "db: "+errAuthor.Error(), http.StatusInternalServerError)
		return
	}

	srv.listAuthors(w, r)
}

func (srv *service) listAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authors, err := srv.storage.Author.Query().All(ctx)
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := assets.ExecuteTemplate(w, "authors", authors); err != nil {
		log.Printf("ERROR: template: %v", err)
		return
	}
}

func logMW(next http.Handler) http.Handler {
	var handle http.HandlerFunc = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("HTTP: %s %s", req.Method, req.URL)
		next.ServeHTTP(rw, req)
	}
	return handle
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
