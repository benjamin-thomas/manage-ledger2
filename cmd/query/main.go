package main

import (
	"fmt"
	"log"
	"os"

	"github.com/benjamin-thomas/manage-ledger2"
	"github.com/olekukonko/tablewriter"

	_ "github.com/lib/pq"
)

func init() {
	if os.Getenv("DEBUG") == "1" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

}

func main() {
	db := manageLedger2.LoadDB()

	rows, err := db.Query(`
		SELECT a.name AS account_name
			 , SUM(p.cents) AS sum_cents
		FROM accounts AS a
	 INNER
		JOIN postings AS p
	 USING (account_id)
	 WHERE a.name LIKE 'Expenses%'
	 GROUP BY account_id
	 ORDER BY 2 DESC
	`)
	if err != nil {
		log.Fatal(err)
	}

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
			log.Fatal(err)
		}
		total := float64(sumCents) / 100
		table.Append([]string{accountName, fmt.Sprintf("%.2f€", total)})
	}
	if err := rows.Close(); err != nil {
		log.Fatal(err)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	noEmptyStringFooterBugRightBorderMissing := " "
	table.SetFooter([]string{"hledger bal Expenses --flat -S", noEmptyStringFooterBugRightBorderMissing})

	table.Render()
}
