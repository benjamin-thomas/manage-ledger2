package manageLedger2

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/k0kubun/pp"
)

var abort = Abort

var accountsSummaryStmt *sql.Stmt

func init() {
	db := LoadDB()
	accountsSummaryStmt = prepareAccountsSummary2(db)
}

func mustPrepare(db *sql.DB, name, qry string) *sql.Stmt {
	stmt, err := db.Prepare(qry)
	if err != nil {
		abort(err, fmt.Sprintf("Could not prepare: '%s'", name))
	}
	return stmt
}

func prepareAccountsSummary2(db *sql.DB) *sql.Stmt {
	return mustPrepare(db, "SummarizeAccounts",
		"SELECT * FROM summarize_accounts($1, $2, $3, $4)",
	)
}

// SummarizeAccounts aggregates cents by account, with cumulative total, and date filters
func SummarizeAccounts(includeAccount string, excludeAccount *string, from, to *time.Time) *sql.Rows {
	if os.Getenv("DEBUG") == "1" {
		pp.Printf("DEBUG: includeAccount: %v\n", includeAccount)
		pp.Printf("DEBUG: excludeAccount: %v\n", excludeAccount)
		pp.Printf("DEBUG: from: %v\n", from)
		pp.Printf("DEBUG: to: %v\n", to)
	}

	rows, err := accountsSummaryStmt.Query(includeAccount, excludeAccount, from, to)
	if err != nil {
		abort(err, "Could not execute summary")
	}
	return rows
}
