CREATE TABLE IF NOT EXISTS product_categories (
    product_id UUID NOT NULL,
    category_id UUID NOT NULL,
    PRIMARY KEY (product_id, category_id),
    
    CONSTRAINT fk_pc_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    CONSTRAINT fk_pc_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS specification_attributes (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE
);



CREATE TABLE IF NOT EXISTS product_specifications (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    attribute_id UUID NOT NULL,
    value VARCHAR(255) NOT NULL,
    
    CONSTRAINT fk_ps_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    CONSTRAINT fk_ps_attribute FOREIGN KEY (attribute_id) REFERENCES specification_attributes (id) ON DELETE CASCADE
);