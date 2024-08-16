package utils

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	dbHost := GetEnv("DB_HOST", "localhost")
	dbUser := GetEnv("DB_USER", "postgres")
	dbPassword := GetEnv("DB_PASSWORD", "postgres")
	dbName := GetEnv("DB_NAME", "auth--service")

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	return sql.Open("postgres", connStr)
}
