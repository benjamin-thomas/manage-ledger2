DROP FUNCTION IF EXISTS summarize_accounts;

CREATE FUNCTION summarize_accounts(
    include VARCHAR(100)
  , exclude VARCHAR(100) = NULL
  , from_ts TIMESTAMP WITHOUT TIME ZONE = NULL
  , to_ts TIMESTAMP WITHOUT TIME ZONE = NULL
)

RETURNS TABLE (
    account_name VARCHAR(100)
  , account_cents BIGINT
  , total_cents NUMERIC
)

AS $BODY$

  SELECT *
       , SUM(x.account_cents) OVER (ORDER BY x.account_cents DESC) AS total_cents
    FROM (
      SELECT a.name AS account_name
           , SUM(p.cents) AS account_cents
        FROM accounts AS a
       INNER JOIN postings AS p USING (account_id)
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

    GROUP BY account_id
         ) AS x
   ORDER BY x.account_cents DESC

$BODY$

LANGUAGE SQL STABLE
