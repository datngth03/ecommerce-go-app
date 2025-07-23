// internal/payment/infrastructure/repository/postgres_payment_repo.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/payment/domain"
)

// PostgreSQLPaymentRepository implements the domain.PaymentRepository interface
// for PostgreSQL database operations.
type PostgreSQLPaymentRepository struct {
	db *sql.DB
}

// NewPostgreSQLPaymentRepository creates a new instance of PostgreSQLPaymentRepository.
func NewPostgreSQLPaymentRepository(db *sql.DB) *PostgreSQLPaymentRepository {
	return &PostgreSQLPaymentRepository{db: db}
}

// Save creates a new payment record or updates an existing one in PostgreSQL.
func (r *PostgreSQLPaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	// Check if payment exists to decide between INSERT or UPDATE
	existingPayment, err := r.FindByID(ctx, payment.ID)
	if err != nil && err.Error() != "payment not found" {
		return fmt.Errorf("failed to check existing payment: %w", err)
	}

	if existingPayment != nil {
		// Payment exists, perform UPDATE
		query := `
            UPDATE payments
            SET order_id = $1, user_id = $2, amount = $3, currency = $4, status = $5, payment_method = $6, transaction_id = $7, updated_at = $8
            WHERE id = $9`
		_, err = r.db.ExecContext(ctx, query,
			payment.OrderID, payment.UserID, payment.Amount, payment.Currency, payment.Status, payment.PaymentMethod, payment.TransactionID, time.Now(), payment.ID)
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
		return nil
	}

	// Payment does not exist, perform INSERT
	query := `
        INSERT INTO payments (id, order_id, user_id, amount, currency, status, payment_method, transaction_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = r.db.ExecContext(ctx, query,
		payment.ID, payment.OrderID, payment.UserID, payment.Amount, payment.Currency, payment.Status, payment.PaymentMethod, payment.TransactionID, payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		// Handle unique constraint violation for transaction_id
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_payments_transaction_id"` ||
			err.Error() == `pq: duplicate key value violates unique constraint "payments_transaction_id_key"` {
			return errors.New("payment with this transaction ID already exists")
		}
		return fmt.Errorf("failed to save new payment: %w", err)
	}
	return nil
}

// FindByID retrieves a payment by its ID from PostgreSQL.
func (r *PostgreSQLPaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	payment := &domain.Payment{}
	query := `SELECT id, order_id, user_id, amount, currency, status, payment_method, transaction_id, created_at, updated_at FROM payments WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.PaymentMethod, &payment.TransactionID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to find payment by ID: %w", err)
	}
	return payment, nil
}

// FindByOrderID retrieves payments associated with a specific order ID from PostgreSQL.
func (r *PostgreSQLPaymentRepository) FindByOrderID(ctx context.Context, orderID string) ([]*domain.Payment, error) {
	var payments []*domain.Payment
	query := `SELECT id, order_id, user_id, amount, currency, status, payment_method, transaction_id, created_at, updated_at FROM payments WHERE order_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments by order ID: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		payment := &domain.Payment{}
		err := rows.Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
			&payment.Status, &payment.PaymentMethod, &payment.TransactionID, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment row by order ID: %w", err)
		}
		payments = append(payments, payment)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error for payments by order ID: %w", err)
	}
	return payments, nil
}

// FindByTransactionID retrieves a payment by its transaction ID from the gateway from PostgreSQL.
func (r *PostgreSQLPaymentRepository) FindByTransactionID(ctx context.Context, transactionID string) (*domain.Payment, error) {
	payment := &domain.Payment{}
	query := `SELECT id, order_id, user_id, amount, currency, status, payment_method, transaction_id, created_at, updated_at FROM payments WHERE transaction_id = $1`
	row := r.db.QueryRowContext(ctx, query, transactionID)

	err := row.Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.PaymentMethod, &payment.TransactionID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to find payment by transaction ID: %w", err)
	}
	return payment, nil
}

// FindAll retrieves a list of payments based on filters and pagination from PostgreSQL.
func (r *PostgreSQLPaymentRepository) FindAll(ctx context.Context, userID, orderID, status string, limit, offset int32) ([]*domain.Payment, int32, error) {
	var payments []*domain.Payment
	var totalCount int32

	baseQuery := `SELECT id, order_id, user_id, amount, currency, status, payment_method, transaction_id, created_at, updated_at FROM payments`
	countQuery := `SELECT COUNT(*) FROM payments`
	args := []interface{}{}
	whereClause := ""
	argCounter := 1

	if userID != "" {
		whereClause += fmt.Sprintf(" WHERE user_id = $%d", argCounter)
		args = append(args, userID)
		argCounter++
	}
	if orderID != "" {
		if whereClause == "" {
			whereClause += fmt.Sprintf(" WHERE order_id = $%d", argCounter)
		} else {
			whereClause += fmt.Sprintf(" AND order_id = $%d", argCounter)
		}
		args = append(args, orderID)
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

	// Get total count first
	err := r.db.QueryRowContext(ctx, countQuery+whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total payment count: %w", err)
	}

	// Add pagination
	query := baseQuery + whereClause + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		payment := &domain.Payment{}
		err := rows.Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
			&payment.Status, &payment.PaymentMethod, &payment.TransactionID, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment row: %w", err)
		}
		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return payments, totalCount, nil
}

// Delete removes a payment record from PostgreSQL by its ID.
func (r *PostgreSQLPaymentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM payments WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("payment not found")
	}
	return nil
}
