package manageLedger2

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // postgres
)

// LoadDB tests and returns a DB
func LoadDB() *sql.DB {
	abort := Abort
	conninfo := fmt.Sprintf("sslmode=disable host=%s dbname=postgres", os.Getenv("PGHOST"))
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		abort(err, "Could not open db")
	}

	if err := db.Ping(); err != nil {
		abort(err, "Could not ping the database")
	}

	return db
}
