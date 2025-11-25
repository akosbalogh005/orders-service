-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    order_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on customer_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);

-- Create index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- Create index on order_time for time-based queries
CREATE INDEX IF NOT EXISTS idx_orders_order_time ON orders(order_time);

-- Create new idempotency_keys table with endpoint name, scheme, response, and validity
CREATE TABLE idempotency_keys (
    endpoint_name VARCHAR(255) NOT NULL,
    endpoint_scheme VARCHAR(10) NOT NULL,
    key VARCHAR(255) NOT NULL,
    response JSONB NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (endpoint_name, endpoint_scheme, key)
);

-- Create index on valid_to for efficient cleanup queries
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_valid_to ON idempotency_keys(valid_to);
