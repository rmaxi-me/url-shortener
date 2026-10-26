package main

import (
	"fmt"
	"net/http"
	"os"
	"rmaxi-me/url-shortner/src/config"
	"rmaxi-me/url-shortner/src/handlers"
	"rmaxi-me/url-shortner/src/models"
	"rmaxi-me/url-shortner/src/stores"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDb(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
		cfg.PostgresIp,
		cfg.PostgresUsername,
		cfg.PostgresPassword,
		cfg.PostgresDb,
		cfg.PostgresPort,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
}

func main() {
	fmt.Println("=== URL Shortener server ===")

	if err := godotenv.Load(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't load env file: %v\n", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't load env file: %v\n", err)
		os.Exit(1)
	}

	pool, err := initDb(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Cannot create DB: %v\n", err)
		os.Exit(1)
	}
	if err := pool.AutoMigrate(&models.ShortUrl{}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Auto migrate failed: %v\n", err)
		os.Exit(1)
	}

	if cfg.UrlIdLength <= 0 {
		_, _ = fmt.Fprintf(os.Stderr, "URL_ID_LENGTH must be a positive integer\n")
		os.Exit(1)
	}

	mux := handlers.RegisterRoutes(stores.NewPostgresStore(cfg.UrlIdLength, pool))
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%v", cfg.ServerPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while running the server: %v\n", err)
	}
}
