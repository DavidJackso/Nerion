package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"go-template/internal/config"
	"go-template/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DB.DSN)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}

	if err := goose.Up(db, "."); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	log.Println("migrations applied")
}
