package data

import (
	"errors"
	"time"
)

type ShortUrl struct {
	ID         int64     `db:"id"`
	Source     string    `db:"source"`
	Target     string    `db:"target"`
	DeleteKey  string    `db:"delete_key"`
	ServeCount int64     `db:"serve_count"`
	CreatedAt  time.Time `db:"created_at"`
}

const ShortUrlColumns = `id, source, target, delete_key, serve_count, created_at`

var ErrSourceAlreadyExists = errors.New("source url already exists")
var ErrSourceNotFound = errors.New("source not found")
var ErrInvalidDeleteKey = errors.New("invalid deletion key")
