package stores

import (
	"context"
	"rmaxi-me/url-shortner/src/models"
)

type Store interface {
	Create(ctx context.Context, shortUrl *models.ShortUrl) (*models.ShortUrl, error)
	Resolve(ctx context.Context, source string) (*models.ShortUrl, error)
	Delete(ctx context.Context, source, deleteKey string) error

	GetIdLength() int
}
