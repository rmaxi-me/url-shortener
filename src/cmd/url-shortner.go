package main

import (
	"fmt"
	"net/http"
	"rmaxi-me/url-shortner/src/config"
	"rmaxi-me/url-shortner/src/handlers"
	"time"
)

func main() {
	mux := handlers.RegisterRoutes()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%v", config.ServerPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Printf("Error while running the server: %v\n", err)
	}
}
