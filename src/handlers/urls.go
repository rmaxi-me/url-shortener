package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"rmaxi-me/url-shortner/src"
	"rmaxi-me/url-shortner/src/models"
	"rmaxi-me/url-shortner/src/stores"
	"rmaxi-me/url-shortner/src/usecases"
	"strings"
)

type UrlRouter struct {
	UrlUsecases *usecases.UrlUsecases
}

func NewUrlRouter(mux *http.ServeMux, store stores.Store) *UrlRouter {
	urlRouter := UrlRouter{UrlUsecases: usecases.NewUrlUsecases(store)}

	mux.HandleFunc("POST /{$}", urlRouter.CreateUrl)
	mux.HandleFunc("GET /{id}", urlRouter.VisitUrl)
	mux.HandleFunc("DELETE /{id}", urlRouter.DeleteUrl)

	return &urlRouter
}

func RegisterRoutes(store stores.Store) *http.ServeMux {
	mux := http.NewServeMux()
	_ = NewUrlRouter(mux, store)

	return mux
}

func (urlRouter *UrlRouter) CreateUrl(writer http.ResponseWriter, request *http.Request) {
	if !strings.Contains(request.Header.Get("Content-Type"), "application/json") {
		src.WriteJSON(writer, http.StatusUnsupportedMediaType, map[string]any{
			"error": "Invalid Content-Type",
		})
		return
	}

	decoder := json.NewDecoder(request.Body)
	var createUrl models.CreateShortUrlBody
	err := decoder.Decode(&createUrl)
	if err != nil {
		src.WriteJSON(writer, http.StatusBadRequest, map[string]any{
			"error": "Invalid body",
		})
		return
	}
	shortUrl, err := urlRouter.UrlUsecases.CreateUrl(createUrl.Url)
	if shortUrl == nil || err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidUrl):
			src.WriteJSON(writer, http.StatusBadRequest, map[string]any{
				"error": "invalid url",
			})
			return
		case errors.Is(err, models.ErrSourceAlreadyExists):
			src.WriteJSON(writer, http.StatusInternalServerError, map[string]any{
				"error": "internal server error",
			})
			return
		default:
			src.WriteJSON(writer, http.StatusInternalServerError, map[string]any{
				"error": "internal server error",
			})
			return
		}
	}

	writer.Header().Set("Location", fmt.Sprintf("/%v", shortUrl.Source))
	src.WriteJSON(writer, http.StatusCreated, map[string]any{
		"id":         shortUrl.Source,
		"delete_key": shortUrl.DeleteKey,
	})
}

func (urlRouter *UrlRouter) VisitUrl(writer http.ResponseWriter, request *http.Request) {

}

func (urlRouter *UrlRouter) DeleteUrl(writer http.ResponseWriter, request *http.Request) {

}
