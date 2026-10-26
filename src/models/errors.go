package models

import "errors"

var ErrSourceAlreadyExists = errors.New("source url already exists")
var ErrSourceNotFound = errors.New("source not found")
var ErrInvalidDeleteKey = errors.New("invalid deletion key")
var ErrInvalidUrl = errors.New("invalid url")
