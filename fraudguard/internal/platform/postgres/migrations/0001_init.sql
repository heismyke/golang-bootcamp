CREATE TABLE accounts (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    external_ref text        NOT NULL UNIQUE,
    home_country char(2)     NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
    id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key   text        NOT NULL UNIQUE,
    account_id        uuid        NOT NULL REFERENCES accounts (id),
    -- Money is stored as an integer count of the currency's minor unit
    -- (cents for USD, yen for JPY). Never use floats for money.
    amount_minor      bigint      NOT NULL CHECK (amount_minor > 0),
    currency          char(3)     NOT NULL,
    merchant_id       text        NOT NULL,
    merchant_category text        NOT NULL,
    country           char(2)     NOT NULL,
    occurred_at       timestamptz NOT NULL,
    risk_score        smallint    NOT NULL CHECK (risk_score BETWEEN 0 AND 100),
    decision          text        NOT NULL CHECK (decision IN ('approve', 'review', 'decline')),
    reasons           text[]      NOT NULL DEFAULT '{}',
    created_at        timestamptz NOT NULL DEFAULT now()
);

-- Serves both the per-account history aggregates used for scoring and the
-- keyset-paginated transaction listing.
CREATE INDEX transactions_account_occurred_idx
    ON transactions (account_id, occurred_at DESC, id DESC);
