-- Generates realistic-looking data for SQL practice: ~1,000 accounts and
-- ~200,000 transactions over the last 90 days. Safe to run more than once.
-- The API must have started at least once so the tables exist.

\timing on

INSERT INTO accounts (external_ref, home_country, created_at)
SELECT 'seed_user_' || g,
       (ARRAY['JP', 'JP', 'JP', 'US', 'GB', 'SG', 'KR'])[1 + floor(random() * 7)::int],
       now() - random() * interval '365 days'
FROM generate_series(1, 1000) AS g
ON CONFLICT (external_ref) DO NOTHING;

WITH acc AS (
    SELECT id, home_country, row_number() OVER (ORDER BY id) AS rn
    FROM accounts
    WHERE external_ref LIKE 'seed_user_%'
),
gen AS (
    SELECT g,
           random() AS score_roll,
           random() AS country_roll
    FROM generate_series(1, 200000) AS g
)
INSERT INTO transactions (idempotency_key, account_id, amount_minor, currency, merchant_id,
                          merchant_category, country, occurred_at, risk_score, decision, reasons)
SELECT 'seed_' || gen.g,
       acc.id,
       -- Skewed toward small amounts, like real card spend.
       (100 + floor(random() * random() * 300000))::bigint,
       'JPY',
       'm_' || (1 + floor(random() * 500))::int,
       (ARRAY['grocery', 'restaurants', 'electronics', 'travel', 'fuel',
              'gambling', 'crypto', 'gift_cards'])[1 + floor(random() * 8)::int],
       CASE WHEN gen.country_roll < 0.9 THEN acc.home_country
            ELSE (ARRAY['US', 'BR', 'NG', 'RU', 'VN', 'FR'])[1 + floor(random() * 6)::int] END,
       now() - random() * interval '90 days',
       floor(gen.score_roll * 100)::int,
       CASE WHEN gen.score_roll >= 0.8 THEN 'decline'
            WHEN gen.score_roll >= 0.4 THEN 'review'
            ELSE 'approve' END,
       '{}'
FROM gen
JOIN acc ON acc.rn = 1 + (gen.g * 7919) % (SELECT count(*) FROM acc)
ON CONFLICT (idempotency_key) DO NOTHING;

ANALYZE accounts;
ANALYZE transactions;

SELECT (SELECT count(*) FROM accounts) AS accounts,
       (SELECT count(*) FROM transactions) AS transactions;
