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
	"github.com/benjamin-thomas/manage-ledger2"

	_ "github.com/lib/pq"
	"gopkg.in/cheggaaa/pb.v1"
)

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
		log.Fatalf("Could not read file: '%s'", path)
	}

	err = json.Unmarshal(bs, &txs)
	if err != nil {
		log.Fatalln("Could not unmarshal json")
	}
	return txs
}

func createTables(db *sql.DB) {
	_, err := db.Exec(`
	DROP TABLE IF EXISTS postings;
	DROP TABLE IF EXISTS transactions;
	DROP TABLE IF EXISTS accounts;

	CREATE TABLE accounts (
			account_id SERIAL PRIMARY KEY
		, name VARCHAR(100) NOT NULL UNIQUE CHECK (TRIM(name) != '')
	);

	CREATE TABLE transactions (
		  transaction_id SERIAL PRIMARY KEY
		, guid UUID NOT NULL UNIQUE
		, descr VARCHAR(255) NOT NULL
		, comment TEXT NULL CHECK (TRIM(comment) != '')
	);

	CREATE TABLE postings (
		  transaction_id INTEGER NOT NULL REFERENCES transactions(transaction_id)
		, timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL
    , account_id INT NOT NULL REFERENCES accounts(account_id)
		, cents INTEGER NOT NULL
		, comment TEXT NULL CHECK (TRIM(comment) != '')
		, mid_comment TEXT NULL CHECK (TRIM(mid_comment) != '')
		, ofx_id VARCHAR(100) NULL UNIQUE CHECK (TRIM(ofx_id) != '')
	)
	`)
	if err != nil {
		log.Fatal("Could not create tables")
	}
}

func prepare(db *sql.DB, name, qry string) *sql.Stmt {
	stmt, err := db.Prepare(qry)
	if err != nil {
		log.Fatalf("Could not prepare: '%s'", name)
	}

	return stmt

}

func prepareInsertTx(db *sql.DB) *sql.Stmt {
	return prepare(db, "insertTxStmt", `
		INSERT INTO transactions (guid, descr, comment)
										VALUES (  $1,    $2,      $3)
										RETURNING transaction_id
	`)
}

func insertTx(insertTxStmt *sql.Stmt, guid, descr string, comment *string) int {
	var transactionID int
	err := insertTxStmt.QueryRow(guid, descr, comment).Scan(&transactionID)
	if err != nil {
		log.Fatalf("Could not insert transaction: '%s'", guid)
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
		log.Fatalln("Could not insert posting")
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
		log.Fatalf("Unexpected failure when finding account by name")
	}

	err = insertAccountStmt.QueryRow(accountName).Scan(&accountID)
	if err != nil {
		log.Fatalln("Could not insert account: '%s'", accountName)
	}
	return accountID
}

func main() {
	file := flag.String("file", "", "JSON file")
	flag.Parse()

	if *file == "" {
		log.Fatalln("Must give -file FILENAME")
	}
	txs := loadJSON(*file)

	db := manageLedger2.LoadDB()

	createTables(db)

	insertTxStmt := prepareInsertTx(db)
	insertPostingStmt := prepareInsertPosting(db)
	insertAccountStmt := prepareInsertAccount(db)
	findAccountByNameStmt := prepareFindAccountByName(db)

	bar := pb.New(len(txs))
	bar.SetRefreshRate(time.Millisecond * 100)
	bar.SetWidth(80)
	bar.Start()
	for _, tx := range txs {
		bar.Increment()
		transactionID := insertTx(insertTxStmt, tx.GUID, tx.Descr, tx.Comment)

		for _, p := range tx.Postings {
			accountID := findOrCreateAccountByName(findAccountByNameStmt, insertAccountStmt, p.Account)
			insertPosting(insertPostingStmt, transactionID, p.Timestamp, accountID, p.Cents, p.Comment, p.MidComment, p.OfxID)
		}
	}
	bar.Finish()

	fmt.Println("Import completed successfully!")
}
