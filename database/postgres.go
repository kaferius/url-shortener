package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresDB() *sql.DB {
	connStr := "host=" + os.Getenv("DB_HOST") + " port=5432 user=postgres password=postgres dbname=mydb sslmode=disable"

	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			_, err = db.Exec("CREATE TABLE IF NOT EXISTS links (short TEXT PRIMARY KEY, long TEXT, clicks INTEGER, expires_at TIMESTAMP, created_at TIMESTAMP NOT NULL DEFAULT NOW());")
		}
		if err == nil {
			return db
		}

		time.Sleep(1 * time.Second)
	}

	log.Fatal(err)
	return nil
}
