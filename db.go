package manageLedger2

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func LoadDB() *sql.DB {
	conninfo := fmt.Sprintf("sslmode=disable host=%s dbname=postgres", os.Getenv("PGHOST"))
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.Fatalln("Could not open db")
	}

	return db
}
