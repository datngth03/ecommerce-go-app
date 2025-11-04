package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
	"github.com/google/uuid"
)

type OrderPostgresRepository struct {
	db *sql.DB
}

func NewOrderPostgresRepository(db *sql.DB) *OrderPostgresRepository {
	return &OrderPostgresRepository{db: db}
}

func (r *OrderPostgresRepository) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	order.ID = uuid.New().String()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO orders (id, user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at`

	err = tx.QueryRowContext(ctx, query,
		order.ID, order.UserID, order.Status, order.TotalAmount,
		order.ShippingAddress, order.PaymentMethod,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, product_name, quantity, price, subtotal, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`

	for i := range order.Items {
		order.Items[i].ID = uuid.New().String()
		order.Items[i].OrderID = order.ID
		order.Items[i].Subtotal = float64(order.Items[i].Quantity) * order.Items[i].Price

		_, err = tx.ExecContext(ctx, itemQuery,
			order.Items[i].ID, order.Items[i].OrderID, order.Items[i].ProductID,
			order.Items[i].ProductName, order.Items[i].Quantity, order.Items[i].Price,
			order.Items[i].Subtotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (r *OrderPostgresRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	order := &models.Order{}

	query := `
		SELECT id, user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at
		FROM orders WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
		&order.ShippingAddress, &order.PaymentMethod,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	itemQuery := `
		SELECT id, order_id, product_id, product_name, quantity, price, subtotal, created_at
		FROM order_items WHERE order_id = $1 ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		err = rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.Price, &item.Subtotal, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *OrderPostgresRepository) List(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*models.Order, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`
	args := []interface{}{userID}
	if status != "" {
		countQuery += ` AND status = $2`
		args = append(args, status)
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get orders
	query := `
		SELECT id, user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at
		FROM orders WHERE user_id = $1`
	if status != "" {
		query += ` AND status = $2`
	}
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	orders := []*models.Order{}
	for rows.Next() {
		order := &models.Order{}
		err = rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
			&order.ShippingAddress, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

func (r *OrderPostgresRepository) UpdateStatus(ctx context.Context, id, status string) (*models.Order, error) {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *OrderPostgresRepository) Cancel(ctx context.Context, id string, userID int64) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3 AND status = $4`
	result, err := r.db.ExecContext(ctx, query, models.OrderStatusCancelled, id, userID, models.OrderStatusPending)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("order not found or cannot be cancelled")
	}

	return nil
}

// ConnectPostgres creates a PostgreSQL database connection
func ConnectPostgres(dsn string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool from config
	if maxOpenConns > 0 {
		db.SetMaxOpenConns(maxOpenConns)
	}
	if maxIdleConns > 0 {
		db.SetMaxIdleConns(maxIdleConns)
	}
	// Set connection lifetime
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}
