package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"database/sql"

	lib "github.com/benjamin-thomas/manage-ledger2"

	"gopkg.in/cheggaaa/pb.v1"
)

const (
	flagError = iota + 2
)

var abort = lib.Abort

// Posting represents a sub transaction
type Posting struct {
	Timestamp  time.Time
	Account    string
	Cents      int
	Currency   string
	Comment    *string
	MidComment *string
	OfxID      *string
}

// Transaction contains a slice of postings
type Transaction struct {
	Cleared  bool
	GUID     string
	Descr    string
	Comment  *string
	Postings []Posting
}

func init() {
	if os.Getenv("DEBUG") == "1" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

}

func loadJSON(path string) []Transaction {
	var txs []Transaction
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		abort(err, fmt.Sprintf("Could not read file: '%s'", path))
	}

	err = json.Unmarshal(bs, &txs)
	if err != nil {
		abort(err, "Could not unmarshal json")
	}
	return txs
}

func prepare(db *sql.DB, name, qry string) *sql.Stmt {
	stmt, err := db.Prepare(qry)
	if err != nil {
		abort(err, fmt.Sprintf("Could not prepare: '%s'", name))
	}

	return stmt

}

func prepareInsertTx(db *sql.DB) *sql.Stmt {
	return prepare(db, "insertTxStmt", `
    INSERT INTO transactions (cleared, guid, descr, comment)
                    VALUES   (     $1,   $2,    $3,      $4)
                    RETURNING transaction_id
  `)
}

func insertTx(insertTxStmt *sql.Stmt, cleared bool, guid, descr string, comment *string) int {
	var transactionID int
	err := insertTxStmt.QueryRow(cleared, guid, descr, comment).Scan(&transactionID)
	if err != nil {
		abort(err, fmt.Sprintf("Could not insert transaction: '%s'", guid))
	}
	return transactionID
}

func prepareInsertPosting(db *sql.DB) *sql.Stmt {
	return prepare(db, "insertPosting", `
		INSERT INTO postings (transaction_id, timestamp,  account_id, cents, comment, mid_comment, ofx_id)
									VALUES (            $1,        $2,          $3,    $4,      $5,          $6,     $7)
	`)
}

func insertPosting(insertPostingStmt *sql.Stmt, transactionID int, timestamp time.Time, accountID, cents int, comment, midComment, ofxID *string) {
	_, err := insertPostingStmt.Exec(transactionID, timestamp, accountID, cents, comment, midComment, ofxID)
	if err != nil {
		abort(err, fmt.Sprintf("Could not insert posting: '%s ofx_id=%s'", timestamp, *ofxID))
	}
}

func prepareInsertAccount(db *sql.DB) *sql.Stmt {
	return prepare(db, "insertAccount", `
		INSERT INTO accounts (name)
								 VALUES  (  $1)
								 RETURNING account_id
	`)
}

func prepareFindAccountByName(db *sql.DB) *sql.Stmt {
	return prepare(db, "findAccountByName", `
		SELECT account_id
		  FROM accounts
		 WHERE name = $1 LIMIT 1
	`)
}

func findOrCreateAccountByName(findAccountByNameStmt, insertAccountStmt *sql.Stmt, accountName string) int {
	var accountID int
	err := findAccountByNameStmt.QueryRow(accountName).Scan(&accountID)

	if err == nil {
		return accountID
	}

	if err != sql.ErrNoRows {
		abort(err, "Unexpected failure when finding account by name")
	}

	err = insertAccountStmt.QueryRow(accountName).Scan(&accountID)
	if err != nil {
		abort(err, fmt.Sprintf("Could not insert account: '%s'", accountName))
	}
	return accountID
}

func main() {
	file := flag.String("file", "", "JSON file")
	flag.Parse()

	if *file == "" {
		fmt.Println("Must give -file FILENAME")
		os.Exit(flagError)
	}
	txs := loadJSON(*file)

	db := lib.LoadDB()

	insertTxStmt := prepareInsertTx(db)
	insertPostingStmt := prepareInsertPosting(db)
	insertAccountStmt := prepareInsertAccount(db)
	findAccountByNameStmt := prepareFindAccountByName(db)

	if _, err := db.Exec(`
		DELETE FROM postings;
		DELETE FROM transactions;
		DELETE FROM accounts;
	`); err != nil {
		abort(err, "Could not purge tables")
	}

	bar := pb.New(len(txs)).
		SetRefreshRate(time.Millisecond * 100).
		SetWidth(80).
		Start()

	for _, tx := range txs {
		bar.Increment()
		transactionID := insertTx(insertTxStmt, tx.Cleared, tx.GUID, tx.Descr, tx.Comment)

		for _, p := range tx.Postings {
			accountID := findOrCreateAccountByName(findAccountByNameStmt, insertAccountStmt, p.Account)
			insertPosting(insertPostingStmt, transactionID, p.Timestamp, accountID, p.Cents, p.Comment, p.MidComment, p.OfxID)
		}
	}
	bar.Finish()

	fmt.Println("Import completed successfully!")
}
