package main

import (
	"fmt"
	"net/http"
	"os"
	"rmaxi-me/url-shortner/src/config"
	"rmaxi-me/url-shortner/src/handlers"
	"rmaxi-me/url-shortner/src/models"
	"rmaxi-me/url-shortner/src/stores"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDb() (*gorm.DB, error) {
	fmt.Println("Initializing database...")

	dsn := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
		config.PostgresIp,
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresDb,
		config.PostgresPort,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
}

func main() {
	fmt.Println("=== URL Shortener server ===")

	pool, err := initDb()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Cannot create DB: %v", err)
		os.Exit(1)
	}
	if err := pool.AutoMigrate(&models.ShortUrl{}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Auto migrate failed: %v", err)
		os.Exit(1)
	}

	sourceLen, err := strconv.Atoi(config.UrlIdLength)
	if err != nil || sourceLen <= 0 {
		_, _ = fmt.Fprintf(os.Stderr, "URL_ID_LENGTH must be a positive integer\n")
		os.Exit(1)
	}

	mux := handlers.RegisterRoutes(stores.NewPostgresStore(sourceLen, pool))
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%v", config.ServerPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	fmt.Println("Starting to listen for requests!")

	err = srv.ListenAndServe()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while running the server: %v\n", err)
	}
}
