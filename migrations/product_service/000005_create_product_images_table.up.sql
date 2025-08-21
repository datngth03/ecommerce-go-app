CREATE TABLE IF NOT EXISTS product_images (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    url VARCHAR(255) NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,

    CONSTRAINT fk_product_image FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);