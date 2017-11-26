package manageLedger2

import (
	"database/sql"
)

var abort = Abort

// AccountsSummary aggregates cents by account, filtered via accountName
func AccountsSummary(db *sql.DB, accountName string) (*sql.Rows, int) {
	rows, err := db.Query(`
		SELECT a.name AS account_name
			 , SUM(p.cents) AS sum_cents
		FROM accounts AS a
	 INNER
		JOIN postings AS p
	 USING (account_id)
	 WHERE a.name ~ $1
	 GROUP BY account_id
	 ORDER BY 2 DESC
	`, accountName)
	if err != nil {
		abort(err, "Could not execute aggregate query")
	}

	row := db.QueryRow(`
		SELECT SUM(p.cents) AS sum_cents
		FROM accounts AS a
	 INNER
		JOIN postings AS p
	 USING (account_id)
	 WHERE a.name ~ $1
	`, accountName)

	var total *int
	if err := row.Scan(&total); err != nil {
		abort(err, "Could not sum category total")
	}
	if total == nil {
		return rows, 0
	}

	return rows, *total
}
