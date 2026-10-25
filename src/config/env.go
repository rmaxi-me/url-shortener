package config

import "os"

var ServerPort = os.Getenv("SERVER_PORT")
var UrlIdLength = os.Getenv("URL_ID_LENGTH")

var PostgresUsername = os.Getenv("POSTGRES_USERNAME")
var PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
var PostgresIp = os.Getenv("POSTGRES_IP")
var PostgresPort = os.Getenv("POSTGRES_PORT")
