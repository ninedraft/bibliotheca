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
	"os"
	"os/signal"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/go-chi/chi/v5"
	"github.com/ninedraft/bibliotheca/internal/service"
	"github.com/ninedraft/bibliotheca/storage/ent"

	_ "modernc.org/sqlite"
)

var (
	//go:embed static/*
	static embed.FS

	//go:embed templ/*
	assetsFS embed.FS
	assets   = template.Must(template.ParseFS(assetsFS, "templ/*.html"))
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

	srv := &service.Service{
		Storage: client,
		Static:  static,
		Templ:   assets,
	}
	mux := chi.NewMux().With(logMW)

	srv.BuildRoutes(mux)

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

func logMW(next http.Handler) http.Handler {
	var handle http.HandlerFunc = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("HTTP: %s %s", req.Method, req.URL)
		next.ServeHTTP(rw, req)
	}
	return handle
}
