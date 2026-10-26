package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort       int
	UrlIdLength      int
	PostgresUsername string
	PostgresPassword string
	PostgresIp       string
	PostgresPort     int
	PostgresDb       string
}

func LoadConfig() (*Config, error) {
	urlIdLength, err := strconv.Atoi(os.Getenv("URL_ID_LENGTH"))
	if err != nil {
		return nil, err
	}
	serverPort, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		return nil, err
	}
	postgresPort, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		return nil, err
	}
	return &Config{
		ServerPort:       serverPort,
		UrlIdLength:      urlIdLength,
		PostgresUsername: os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresIp:       os.Getenv("POSTGRES_IP"),
		PostgresPort:     postgresPort,
		PostgresDb:       os.Getenv("POSTGRES_DB"),
	}, nil
}
