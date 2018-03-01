DROP FUNCTION IF EXISTS account_for_relative_month;

CREATE FUNCTION account_for_relative_month (
    account_include_re VARCHAR
  , account_exclude_re VARCHAR
  , relative_month INTEGER
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
     *
    FROM summarize_transactions(
        account_include_re
      , account_exclude_re
      , (DATE_TRUNC('month', current_date) + (relative_month || ' month')::INTERVAL)::DATE
      , (DATE_TRUNC('month', current_date + '1 month'::INTERVAL) + (relative_month || ' month')::INTERVAL)::DATE
    );

$BODY$

LANGUAGE SQL STABLE
