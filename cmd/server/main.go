package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/neizhmak/avito-review-service/internal/service"
	"github.com/neizhmak/avito-review-service/internal/storage/postgres"
	"github.com/neizhmak/avito-review-service/internal/transport/rest"
)

func main() {
	dbConnStr := os.Getenv("DB_CONNECTION_STRING")
	if dbConnStr == "" {
		dbConnStr = "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	// initialize storages
	teamStorage := postgres.NewTeamStorage(db)
	userStorage := postgres.NewUserStorage(db)
	prStorage := postgres.NewPullRequestStorage(db)

	// initialize service
	prService := service.NewPRService(prStorage, userStorage, teamStorage, db)

	// initialize handler (HTTP)
	handler := rest.NewHandler(prService)

	// start server
	log.Printf("Starting server on :%s", port)

	if err := http.ListenAndServe(":"+port, handler.InitRouter()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
