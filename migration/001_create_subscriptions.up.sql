CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    start_date DATE NOT NULL,
    end_date DATE
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);