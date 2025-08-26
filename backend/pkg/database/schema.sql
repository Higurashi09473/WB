CREATE TABLE IF NOT EXISTS delivery (
    order_uid VARCHAR(255) PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    zip TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    region TEXT NOT NULL,
    email TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payment (
    transaction VARCHAR(255) PRIMARY KEY,
    request_id TEXT,
    currency TEXT NOT NULL,
    provider TEXT NOT NULL,
    amount INTEGER NOT NULL,
    payment_dt BIGINT NOT NULL,
    bank TEXT NOT NULL,
    delivery_cost INTEGER NOT NULL,
    goods_total INTEGER NOT NULL,
    custom_fee INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number TEXT NOT NULL,
    entry TEXT NOT NULL,
    delivery_uid VARCHAR(255) NOT NULL,
    payment_transaction VARCHAR(255) NOT NULL,
    locale TEXT NOT NULL,
    internal_signature TEXT NOT NULL,
    customer_id TEXT NOT NULL,
    delivery_service TEXT NOT NULL,
    shardkey TEXT NOT NULL,
    sm_id INTEGER NOT NULL,
    date_created TIMESTAMP WITH TIME ZONE NOT NULL,
    oof_shard TEXT NOT NULL,
    FOREIGN KEY (delivery_uid) REFERENCES delivery(order_uid) ON DELETE CASCADE,
    FOREIGN KEY (payment_transaction) REFERENCES payment(transaction) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) NOT NULL,
    chrt_id INTEGER NOT NULL,
    track_number TEXT NOT NULL,
    price INTEGER NOT NULL,
    rid TEXT NOT NULL,
    name TEXT NOT NULL,
    sale INTEGER NOT NULL,
    size TEXT NOT NULL,
    total_price INTEGER NOT NULL,
    nm_id INTEGER NOT NULL,
    brand TEXT NOT NULL,
    status INTEGER NOT NULL,
    FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
);