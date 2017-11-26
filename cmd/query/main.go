package main

import (
	"fmt"
	"log"
	"os"

	lib "github.com/benjamin-thomas/manage-ledger2"
	"github.com/olekukonko/tablewriter"
)

var abort = lib.Abort

func init() {
	if os.Getenv("DEBUG") == "1" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

}

func main() {
	db := lib.LoadDB()
	if len(os.Args) != 2 {
		abort(nil, "Give account filter (ex: go run ./cmd/query/main.go Expenses.+Presents)")
	}
	accountName := os.Args[1]
	rows, totalCents := lib.AccountsSummary(db, accountName)

	table := tablewriter.NewWriter(os.Stdout)
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
	table.SetFooter([]string{fmt.Sprintf("hledger bal %s --flat -S", accountName), fmt.Sprintf("%.2f€", mainTotal)})

	table.Render()
}
