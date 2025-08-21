CREATE TABLE IF NOT EXISTS product_variants (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL UNIQUE,
    price DECIMAL(12,2) NOT NULL,
    original_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(5,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

);

CREATE TABLE IF NOT EXISTS variant_attributes (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS variant_attribute_values (
    id UUID PRIMARY KEY,
    attribute_id UUID NOT NULL,
    value VARCHAR(100) NOT NULL,

);

CREATE TABLE IF NOT EXISTS product_variant_options (
    variant_id UUID NOT NULL,
    attribute_value_id UUID NOT NULL,
    PRIMARY KEY (variant_id, attribute_value_id),
    
);