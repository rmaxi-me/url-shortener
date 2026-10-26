package models

import "gorm.io/gorm"

type ShortUrl struct {
	gorm.Model

	Source     string `gorm:"not null;uniqueIndex"`
	Target     string `gorm:"not null"`
	DeleteKey  string `gorm:"not null"`
	ServeCount int64  `gorm:"not null;default:0"`
}

type CreateShortUrlBody struct {
	Url string `json:"url"`
}
