package main

import (
	"database/sql"
	"fmt"
	"net/http"

	lib "github.com/benjamin-thomas/manage-ledger2"
	"github.com/olekukonko/tablewriter"
)

var db *sql.DB
var abort = lib.Abort

func init() {
	db = lib.LoadDB()
}

// http://localhost:8080/Expenses.+Presents
func rootHandler(w http.ResponseWriter, r *http.Request) {
	accountFilter := r.URL.Path[1:]
	rows, totalCents := lib.AccountsSummary(db, accountFilter)

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"ACCOUNT", "TOTAL"})
	table.SetColWidth(70)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})

	for rows.Next() {
		var (
			accountName string
			sumCents    int
		)
		if err := rows.Scan(&accountName, &sumCents); err != nil {
			abort(err, "Could not scan account")
		}
		total := float64(sumCents) / 100
		table.Append([]string{accountName, fmt.Sprintf("%.2f€", total)})
	}
	if err := rows.Close(); err != nil {
		abort(err, "Could not close rows")
	}

	if err := rows.Err(); err != nil {
		abort(err, "Rows has errors")
	}

	mainTotal := float64(totalCents) / 100

	table.SetAutoFormatHeaders(false)
	table.SetFooter([]string{fmt.Sprintf("hledger bal %s --flat -S", accountFilter), fmt.Sprintf("%.2f€", mainTotal)})

	table.Render()

	fmt.Fprintf(w, "\naccountFilter=%s", accountFilter)

}

func main() {
	http.HandleFunc("/", rootHandler)

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		abort(err, "Failed to start server")
	}

	if err := db.Close(); err != nil {
		abort(err, "Could not close DB")
	}
}
