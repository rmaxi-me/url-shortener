package stores

import (
	"context"
	"errors"
	"fmt"
	"rmaxi-me/url-shortner/src/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresStore struct {
	db           *gorm.DB
	SourceLength int
}

func NewPostgresStore(sourceLength int, pool *gorm.DB) *PostgresStore {
	return &PostgresStore{db: pool, SourceLength: sourceLength}
}

func (postgres *PostgresStore) Create(ctx context.Context, shortUrl *models.ShortUrl) (*models.ShortUrl, error) {
	err := gorm.G[models.ShortUrl](postgres.db).Create(ctx, shortUrl)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, models.ErrSourceAlreadyExists
	}
	if err != nil {
		return nil, fmt.Errorf("create url: %w", err)
	}
	return shortUrl, err
}

func (postgres *PostgresStore) Resolve(ctx context.Context, source string) (*models.ShortUrl, error) {
	var shortUrl models.ShortUrl

	result := postgres.db.WithContext(ctx).
		Model(&shortUrl).
		Clauses(clause.Returning{}).
		Where("source = ?", source).
		Update("serve_count", gorm.Expr("serve_count + 1"))

	if result.Error != nil {
		return nil, fmt.Errorf("resolve url: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, models.ErrSourceNotFound
	}
	return &shortUrl, nil
}

func (postgres *PostgresStore) Delete(ctx context.Context, source, deleteKey string) error {
	var shortUrl models.ShortUrl

	result := postgres.db.Where("source = ?", source).First(&shortUrl)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.ErrSourceNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("resolve url: %w", result.Error)
	}

	if deleteKey != shortUrl.DeleteKey {
		return models.ErrInvalidDeleteKey
	}

	result = postgres.db.Where("source = ?", source).Delete(&models.ShortUrl{})
	if result.Error != nil {
		return fmt.Errorf("delete url: %w", result.Error)
	}
	return nil
}

func (postgres *PostgresStore) GetIdLength() int {
	return postgres.SourceLength
}
