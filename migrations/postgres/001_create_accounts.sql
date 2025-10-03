CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    balance DECIMAL(20,8) NOT NULL DEFAULT 0.00000000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for minimum balance queries (for transactions)
CREATE INDEX IF NOT EXISTS idx_accounts_balance ON accounts(balance);

-- Create transactions table (audit trail basically)
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    source_account_id INTEGER NOT NULL,
    destination_account_id INTEGER NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (source_account_id) REFERENCES accounts(id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(id)
);