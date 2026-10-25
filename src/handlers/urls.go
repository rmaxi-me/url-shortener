package handlers

import (
	"net/http"
)

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", func(writer http.ResponseWriter, request *http.Request) {
	})

	mux.HandleFunc("GET /{id}", func(writer http.ResponseWriter, request *http.Request) {
	})

	mux.HandleFunc("DELETE /{id}", func(writer http.ResponseWriter, request *http.Request) {
	})

	return mux
}
