// services/product-service/internal/repository/product_repository.go

package repository

import (
	"database/sql"
	"fmt"

	// "github.com/datngth03/ecommerce-go-app/services/product-service/internal/repository/postgres"
	_ "github.com/lib/pq"
)

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(opts *RepositoryOptions) (*Repository, error) {
	db, ok := opts.Database.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("invalid database type, expected *sql.DB")
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{
		Product:  NewProductRepository(db),
		Category: NewCategoryRepository(db),
	}, nil
}

// ConnectPostgres creates a new PostgreSQL database connection
func ConnectPostgres(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// MigrateDatabase runs database migrations
func MigrateDatabase(db *sql.DB) error {
	// Create categories table
	categoryTableQuery := `
		CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL UNIQUE,
			slug VARCHAR(120) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Create products table
	productTableQuery := `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(280) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
			category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
			image_url VARCHAR(500),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	// Create indexes
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);`,
		`CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);`,
		`CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);`,
		`CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);`,
	}

	// Create updated_at trigger function
	triggerFunctionQuery := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`

	// Create triggers
	triggerQueries := []string{
		`DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;`,
		`CREATE TRIGGER update_categories_updated_at 
		 BEFORE UPDATE ON categories 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,
		`DROP TRIGGER IF EXISTS update_products_updated_at ON products;`,
		`CREATE TRIGGER update_products_updated_at 
		 BEFORE UPDATE ON products 
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,
	}

	// Execute migrations
	queries := []string{categoryTableQuery, productTableQuery}
	queries = append(queries, indexQueries...)
	queries = append(queries, triggerFunctionQuery)
	queries = append(queries, triggerQueries...)

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// SeedDatabase seeds the database with initial data
func SeedDatabase(db *sql.DB) error {
	// Check if categories already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing categories: %w", err)
	}

	// If categories exist, skip seeding
	if count > 0 {
		return nil
	}

	// Seed categories
	seedCategories := []struct {
		name string
		slug string
	}{
		{"Electronics", "electronics"},
		{"Clothing", "clothing"},
		{"Books", "books"},
		{"Home & Garden", "home-garden"},
		{"Sports & Outdoors", "sports-outdoors"},
	}

	for _, cat := range seedCategories {
		query := `
			INSERT INTO categories (name, slug) 
			VALUES ($1, $2) 
			ON CONFLICT (name) DO NOTHING
		`
		_, err := db.Exec(query, cat.name, cat.slug)
		if err != nil {
			return fmt.Errorf("failed to seed category %s: %w", cat.name, err)
		}
	}

	return nil
}
