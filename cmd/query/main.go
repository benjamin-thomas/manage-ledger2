package main

import (
	"fmt"
	"log"
	"os"
	"time"

	lib "github.com/benjamin-thomas/manage-ledger2"
	"github.com/olekukonko/tablewriter"
)

var abort = lib.Abort

func init() {
	if os.Getenv("DEBUG") == "1" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

}

func toTime(val *string) *time.Time {
	if val == nil {
		return nil
	}

	var layout string
	// Reference date: Mon Jan 2 15:04:05 MST 2006
	switch len(*val) {
	case 10:
		layout = "2006-01-02"
	case 7:
		layout = "2006-01"
	case 4:
		layout = "2006"
	default:
		log.Panic("Unknown layout for val: " + *val)
	}
	t, err := time.Parse(layout, *val)
	if err != nil {
		abort(err, fmt.Sprintf("Could not parse time: %s", *val))
	}
	return &t
}

func printAccountsSummary(include string, exclude *string, from, to *time.Time) {
	rows := lib.SummarizeAccounts(include, exclude, from, to)
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"ACCOUNT", "ACCOUNT_TOTAL", "TOTAL"})
	table.SetColWidth(70)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})

	for rows.Next() {
		var (
			accountName       string
			accountTotalCents int
			totalCents        int
		)
		if err := rows.Scan(&accountName, &accountTotalCents, &totalCents); err != nil {
			abort(err, "Could not scan")
		}

		accountTotal := float64(accountTotalCents) / 100
		total := float64(totalCents) / 100
		table.Append([]string{accountName, fmt.Sprintf("%.2f€", accountTotal), fmt.Sprintf("%.2f€", total)})

	}

	if err := rows.Close(); err != nil {
		abort(err, "Could not close rows")
	}

	if err := rows.Err(); err != nil {
		abort(err, "Rows has errors")
	}

	table.Render()
	fmt.Printf("\nhledger bal %s --flat -S\n", include)
}

func main() {
	var include string
	var exclude *string
	var from *string
	var to *string

	switch len(os.Args) {
	case 5:
		to = &os.Args[4]
		fallthrough
	case 4:
		from = &os.Args[3]
		if *from == "_" {
			from = nil
		}
		fallthrough
	case 3:
		exclude = &os.Args[2]
		if *exclude == "_" {
			exclude = nil
		}
		fallthrough
	case 2:
		include = os.Args[1]
	case 1:
		abort(nil, "Usage: query includeRegex [excludeRegex from to]")
	}

	printAccountsSummary(include, exclude, toTime(from), toTime(to))
}
