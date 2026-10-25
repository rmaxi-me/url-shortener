package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"rmaxi-me/url-shortner/src/config"
	"rmaxi-me/url-shortner/src/handlers"
	"time"

	"github.com/jackc/pgx/v5"
)

func initDb() *pgx.Conn {
	fmt.Println("Initializing database...")

	url := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v",
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresIp,
		config.PostgresPort,
		config.PostgresDb,
	)
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Database initialized")

	return conn
}

func main() {
	fmt.Println("=== URL Shortener server ===")

	db := initDb()
	defer db.Close(context.Background())

	mux := handlers.RegisterRoutes()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%v", config.ServerPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	fmt.Println("Starting to listen for requests!")

	err := srv.ListenAndServe()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while running the server: %v\n", err)
	}
}
