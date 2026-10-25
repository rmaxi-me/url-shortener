package handlers

import (
	"fmt"
	"net/http"
)

func RegisterRoutes() *http.ServeMux {
	fmt.Println("Registering routes...")

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", func(writer http.ResponseWriter, request *http.Request) {
	})

	mux.HandleFunc("GET /{id}", func(writer http.ResponseWriter, request *http.Request) {
	})

	mux.HandleFunc("DELETE /{id}", func(writer http.ResponseWriter, request *http.Request) {
	})

	fmt.Println("Routes registered.")

	return mux
}
