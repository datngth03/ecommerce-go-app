CREATE TABLE IF NOT EXISTS product_tags (
    product_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    PRIMARY KEY (product_id, tag_id),
    
    CONSTRAINT fk_pt_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    CONSTRAINT fk_pt_tag FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE
);