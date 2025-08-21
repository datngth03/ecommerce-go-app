CREATE TABLE IF NOT EXISTS inventory_transactions (
    id UUID PRIMARY KEY,
    variant_id UUID NOT NULL,
    change_type VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    reference_id UUID,
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
);

