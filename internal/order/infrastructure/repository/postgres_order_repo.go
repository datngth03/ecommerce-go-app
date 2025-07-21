// internal/order/infrastructure/repository/postgres_order_repo.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json" // For marshaling/unmarshaling OrderItem slice
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/order/domain"
)

// PostgreSQLOrderRepository implements the domain.OrderRepository interface
// for PostgreSQL database operations.
type PostgreSQLOrderRepository struct {
	db *sql.DB
}

// NewPostgreSQLOrderRepository creates a new instance of PostgreSQLOrderRepository.
func NewPostgreSQLOrderRepository(db *sql.DB) *PostgreSQLOrderRepository {
	return &PostgreSQLOrderRepository{db: db}
}

// Save creates a new order or updates an existing one in PostgreSQL.
// It handles both the 'orders' table and 'order_items' table.
func (r *PostgreSQLOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil) // Start a transaction
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error, commit manually on success

	// Check if order exists to decide between INSERT or UPDATE
	existingOrder, err := r.FindByID(ctx, order.ID)
	if err != nil && err.Error() != "order not found" {
		return fmt.Errorf("failed to check existing order: %w", err)
	}

	if existingOrder != nil {
		// Order exists, perform UPDATE on 'orders' table
		orderQuery := `
            UPDATE orders
            SET user_id = $1, total_amount = $2, status = $3, shipping_address = $4, updated_at = $5
            WHERE id = $6`
		_, err = tx.ExecContext(ctx, orderQuery,
			order.UserID, order.TotalAmount, order.Status, order.ShippingAddress, time.Now(), order.ID)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		// Delete existing order items for this order
		deleteItemsQuery := `DELETE FROM order_items WHERE order_id = $1`
		_, err = tx.ExecContext(ctx, deleteItemsQuery, order.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing order items: %w", err)
		}
	} else {
		// Order does not exist, perform INSERT into 'orders' table
		orderQuery := `
            INSERT INTO orders (id, user_id, total_amount, status, shipping_address, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err = tx.ExecContext(ctx, orderQuery,
			order.ID, order.UserID, order.TotalAmount, order.Status, order.ShippingAddress, order.CreatedAt, order.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert new order: %w", err)
		}
	}

	// Insert new order items into 'order_items' table
	for _, item := range order.Items {
		itemQuery := `
            INSERT INTO order_items (order_id, product_id, product_name, price, quantity)
            VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.ExecContext(ctx, itemQuery,
			order.ID, item.ProductID, item.ProductName, item.Price, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return tx.Commit() // Commit the transaction
}

// FindByID retrieves an order by its ID from PostgreSQL.
// It also fetches associated order items.
func (r *PostgreSQLOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	order := &domain.Order{}
	orderQuery := `SELECT id, user_id, total_amount, status, shipping_address, created_at, updated_at FROM orders WHERE id = $1`
	row := r.db.QueryRowContext(ctx, orderQuery, id)

	err := row.Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status,
		&order.ShippingAddress, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to find order by ID: %w", err)
	}

	// Fetch order items for this order
	itemsQuery := `SELECT product_id, product_name, price, quantity FROM order_items WHERE order_id = $1`
	rows, err := r.db.QueryContext(ctx, itemsQuery, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items for order %s: %w", order.ID, err)
	}
	defer rows.Close()

	var orderItems []domain.OrderItem
	for rows.Next() {
		item := domain.OrderItem{}
		err := rows.Scan(&item.ProductID, &item.ProductName, &item.Price, &item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item row: %w", err)
		}
		orderItems = append(orderItems, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("order items rows iteration error: %w", err)
	}
	order.Items = orderItems

	return order, nil
}

// FindAll retrieves a list of orders based on filters and pagination from PostgreSQL.
// It also fetches associated order items for each order.
func (r *PostgreSQLOrderRepository) FindAll(ctx context.Context, userID, status string, limit, offset int32) ([]*domain.Order, int32, error) {
	var orders []*domain.Order
	var totalCount int32

	// Build the query dynamically based on filters
	query := `SELECT id, user_id, total_amount, status, shipping_address, created_at, updated_at FROM orders`
	countQuery := `SELECT COUNT(*) FROM orders`
	args := []interface{}{}
	whereClause := ""
	argCounter := 1

	if userID != "" {
		whereClause += fmt.Sprintf(" WHERE user_id = $%d", argCounter)
		args = append(args, userID)
		argCounter++
	}
	if status != "" {
		if whereClause == "" {
			whereClause += fmt.Sprintf(" WHERE status = $%d", argCounter)
		} else {
			whereClause += fmt.Sprintf(" AND status = $%d", argCounter)
		}
		args = append(args, status)
		argCounter++
	}

	query += whereClause
	countQuery += whereClause

	// Get total count first
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total order count: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status,
			&order.ShippingAddress, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order row: %w", err)
		}
		// For simplicity, we're not fetching order items for ListAll.
		// In a real app, you might fetch them in a separate query or join.
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, totalCount, nil
}

// Delete removes an order from PostgreSQL by its ID.
// It also removes associated order items due to ON DELETE CASCADE.
func (r *PostgreSQLOrderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM orders WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
}

// Helper function to marshal/unmarshal OrderItem slice to/from JSONB
// This is an alternative if you were to store items as JSONB in a single column.
// However, we are using a separate 'order_items' table, so this is not directly used
// but kept as an example of JSONB handling.
func marshalOrderItems(items []domain.OrderItem) ([]byte, error) {
	return json.Marshal(items)
}

func unmarshalOrderItems(data []byte) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	if len(data) == 0 {
		return items, nil
	}
	err := json.Unmarshal(data, &items)
	return items, err
}
