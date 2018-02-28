DROP FUNCTION IF EXISTS summarize_transactions;

-- Use as such:
-- SELECT * FROM summarize_transactions('Expenses', NULL, '2018-02-01', '2018-03-01');
-- SELECT * FROM summarize_transactions('Expenses.+Cloth', NULL, '2018-01-01', '2018-02-01');

CREATE FUNCTION summarize_transactions(
    include VARCHAR(100)
  , exclude VARCHAR(100) = NULL
  , from_ts TIMESTAMP WITHOUT TIME ZONE = NULL
  , to_ts TIMESTAMP WITHOUT TIME ZONE = NULL
)

RETURNS TABLE (
    posted_on DATE
  , short_descr CHAR
  , short_comment CHAR
  , account_name CHAR
  , euros NUMERIC(19,2)
  , total_euros NUMERIC(19,2)
  , cents INT
  , total_cents BIGINT
)

AS $BODY$

  SELECT
       y.timestamp::DATE AS posted_on
     , y.descr::CHAR(35) AS short_descr
     , REGEXP_REPLACE(y.comment::CHAR(22), '; ', '') AS short_comment
     , y.account_name::CHAR(35) AS short_account_name
     , (y.cents::NUMERIC/100)::NUMERIC(19,2) AS euros
     , (y.total_cents::NUMERIC/100)::NUMERIC(19,2) AS total_euros
     , y.cents
     , y.total_cents
    FROM (
      SELECT
           x.*
         , SUM(x.cents) OVER(ORDER BY x.timestamp, x.posting_id ASC) AS total_cents  -- don't PARTITION BY, the filter will alter total aggregation

         FROM (

            SELECT
                   p.posting_id
                 , p.timestamp
                 , t.descr
                 , a.name AS account_name
                 , p.cents
                 , p.comment
              FROM accounts AS a
             INNER JOIN postings AS p USING (account_id)
             INNER JOIN transactions AS t USING (transaction_id)
             WHERE a.name ~ $1

               AND (
                CASE $2 IS NULL
                WHEN TRUE THEN
                  TRUE
                ELSE
                  a.name !~ $2
                END
              )

               AND (
                CASE $3 IS NULL
                WHEN TRUE THEN
                  TRUE
                ELSE
                  p.timestamp >= $3
                END
               )

               AND (
                CASE $4 IS NULL
                WHEN TRUE THEN
                  TRUE
                ELSE
                  p.timestamp < $4
                END
               )

        ) AS x
    ) AS y

    ORDER BY y.timestamp, y.posting_id

$BODY$

LANGUAGE SQL STABLE
