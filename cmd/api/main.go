package main

import (
	"log"
	"net/http"

	"investment_committee/internal/api/router"
	"investment_committee/internal/config"
	"investment_committee/internal/db"
	"investment_committee/internal/db/queries"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer conn.Close()

	repo := queries.NewRepository(conn)
	r := router.New(repo, cfg.APIKey)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	log.Printf("listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
