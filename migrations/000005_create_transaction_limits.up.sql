CREATE TABLE transaction_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    max_per_transaction NUMERIC(18,2) NOT NULL DEFAULT 10000.00,
    max_daily_amount NUMERIC(18,2) NOT NULL DEFAULT 50000.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_limits_user ON transaction_limits(user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_limits_global ON transaction_limits((1)) WHERE user_id IS NULL;

INSERT INTO transaction_limits (id, user_id, max_per_transaction, max_daily_amount)
VALUES (gen_random_uuid(), NULL, 10000.00, 50000.00);
