## Setup

./manage/setup_db # then follow instructions


## Query via psql

```console
PAGER="less -S" psql

postgres=# SELECT * FROM summarize_transactions('Expenses:Food', NULL, '2018-01-01', '2018-02-01');
postgres=# SELECT * FROM account_for_relative_month('Expenses:Food', NULL, 0);
postgres=# SELECT posted_on, short_descr, short_comment, euros, total_euros FROM account_for_relative_month('Expenses:Food', NULL, 0);
```
