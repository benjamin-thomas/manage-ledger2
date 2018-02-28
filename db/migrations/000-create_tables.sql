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
    posting_id SERIAL PRIMARY KEY
  , transaction_id INTEGER NOT NULL REFERENCES transactions(transaction_id)
  , timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL
  , account_id INT NOT NULL REFERENCES accounts(account_id)
  , cents INTEGER NOT NULL
  , comment TEXT NULL CHECK (TRIM(comment) != '')
  , mid_comment TEXT NULL CHECK (TRIM(mid_comment) != '')
  , ofx_id VARCHAR(100) NULL UNIQUE CHECK (TRIM(ofx_id) != '')
)
