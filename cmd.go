package main

import (
	"database/sql"
	_ "embed"
	"log/slog"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/caarlos0/env/v11"
	"github.com/kzs0/kokoro"
	"github.com/kzs0/pill_manager/manager"
	"github.com/kzs0/pill_manager/models/db/sqlc"
	"github.com/kzs0/pill_manager/pkg/middleware"
)

type Config struct {
	Koko  kokoro.Config
	Auth0 middleware.Auth0Config
}

// TODO remove/turn into a job
//
//go:embed models/db/schema.sql
var ddl string

func main() {
	config := Config{}
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	ctx, done, err := kokoro.Init(kokoro.WithConfig(config.Koko))
	defer done()
	if err != nil {
		slog.Error("failed to initialize kokoro", slog.Any("err", err))
		panic(err)
	}

	db, err := sql.Open("sqlite3", "manager.db")
	if err != nil {
		panic(err)
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		panic(err)
	}

	queries := sqlc.New(db)

	// rx := repositories.NewPerscriptions()
	// users := repositories.NewUsers()
	// doses := repositories.NewDoses()
	handler := manager.Handler{
		Queries: queries,
	}

	controller := manager.Controller{
		Queries: queries,
		Handler: &handler,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", controller.GetRoot)
	mux.HandleFunc("GET /rx/remaining", controller.GetRemainingDoses)
	mux.HandleFunc("GET /rx/{id}", controller.GetPerscription)
	mux.HandleFunc("POST /rx/taken/{id}", controller.PostTaken)
	mux.HandleFunc("POST /rx", controller.PostPerscription)
	mux.HandleFunc("POST /user", controller.PostUser)
	mux.HandleFunc("OPTIONS /rx", controller.Options)

	// wrappedMux := middleware.HttpOperation(ctx, mux)

	opts := &middleware.CORSOptions{
		Origin:  []string{"*"},
		Methods: []string{"GET", "POST", "OPTIONS"},
		Headers: []string{"Content-Type", "Authorization"},
	}

	userMux := middleware.ObserveNewUsers(mux, queries)
	jwtMux := middleware.EnsureValidToken(userMux, config.Auth0)
	corsMux := middleware.CORS(jwtMux, opts)

	if err := http.ListenAndServe(":8080", corsMux); err != nil {
		slog.Error("server failed", slog.Any("err", err))
		panic(err)
	}
}
