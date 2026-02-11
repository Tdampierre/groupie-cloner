package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func InitDB() (*sql.DB, error) {
	if os.Getenv("DISABLE_DB") == "1" {
		log.Println("DB disabled via DISABLE_DB=1")
		return nil, nil
	}

	dsn := mysql.Config{
		User:                 getenvDefault("DB_USER", "root"),
		Passwd:               getenvDefault("DB_PASS", ""),
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", getenvDefault("DB_HOST", "localhost"), getenvDefault("DB_PORT", "3306")),
		DBName:               getenvDefault("DB_NAME", "groupi_tracker"),
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}

	database, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(30 * time.Minute)

	if err := database.Ping(); err != nil {
		return nil, err
	}

	log.Println("DB connection established")
	return database, nil
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
