package usecases

import (
	"context"
	"errors"
	"net/url"
	"rmaxi-me/url-shortner/src"
	"rmaxi-me/url-shortner/src/models"
	"rmaxi-me/url-shortner/src/stores"
	"strings"
)

const (
	deleteKeyLength     = 32
	maxCreationAttempts = 10
	maxUrlLength        = 2048
)

type UrlUsecases struct {
	UrlStore stores.Store
}

func NewUrlUsecases(store stores.Store) *UrlUsecases {
	return &UrlUsecases{UrlStore: store}
}

func (urlUsecases *UrlUsecases) CreateUrl(target string) (*models.ShortUrl, error) {
	if len(target) > maxUrlLength {
		return nil, models.ErrInvalidUrl
	}

	u, err := url.ParseRequestURI(target)
	if err != nil {
		return nil, models.ErrInvalidUrl
	}
	if scheme := strings.ToLower(u.Scheme); scheme != "http" && scheme != "https" {
		return nil, models.ErrInvalidUrl
	}
	if u.Host == "" {
		return nil, models.ErrInvalidUrl
	}

	for range maxCreationAttempts {
		shortUrl := models.ShortUrl{
			Source:    src.RandHexString(urlUsecases.UrlStore.GetIdLength()),
			Target:    target,
			DeleteKey: src.RandHexString(deleteKeyLength),
		}
		created, err := urlUsecases.UrlStore.Create(context.Background(), &shortUrl)
		if errors.Is(err, models.ErrSourceAlreadyExists) {
			continue
		}
		return created, err
	}
	return nil, models.ErrSourceAlreadyExists
}

func (urlUsecases *UrlUsecases) ResolveId(source string) (*models.ShortUrl, error) {
	return urlUsecases.UrlStore.Resolve(context.Background(), source)
}

func (urlUsecases *UrlUsecases) DeleteUrl(shortUrl models.ShortUrl) {

}
