// internal/shipping/infrastructure/repository/postgres_shipment_repo.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/shipping/domain"
)

// PostgreSQLShipmentRepository implements the domain.ShipmentRepository interface
// for PostgreSQL database operations.
type PostgreSQLShipmentRepository struct {
	db *sql.DB
}

// NewPostgreSQLShipmentRepository creates a new instance of PostgreSQLShipmentRepository.
func NewPostgreSQLShipmentRepository(db *sql.DB) *PostgreSQLShipmentRepository {
	return &PostgreSQLShipmentRepository{db: db}
}

// Save creates a new shipment record or updates an existing one in PostgreSQL.
func (r *PostgreSQLShipmentRepository) Save(ctx context.Context, shipment *domain.Shipment) error {
	// Check if shipment exists to decide between INSERT or UPDATE
	existingShipment, err := r.FindByID(ctx, shipment.ID)
	if err != nil && err.Error() != "shipment not found" {
		return fmt.Errorf("failed to check existing shipment: %w", err)
	}

	if existingShipment != nil {
		// Shipment exists, perform UPDATE
		query := `
            UPDATE shipments
            SET order_id = $1, user_id = $2, shipping_cost = $3, tracking_number = $4, carrier = $5, status = $6, shipping_address = $7, updated_at = $8
            WHERE id = $9`
		_, err = r.db.ExecContext(ctx, query,
			shipment.OrderID, shipment.UserID, shipment.ShippingCost, shipment.TrackingNumber, shipment.Carrier, shipment.Status, shipment.ShippingAddress, time.Now(), shipment.ID)
		if err != nil {
			return fmt.Errorf("failed to update shipment: %w", err)
		}
		return nil
	}

	// Shipment does not exist, perform INSERT
	query := `
        INSERT INTO shipments (id, order_id, user_id, shipping_cost, tracking_number, carrier, status, shipping_address, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = r.db.ExecContext(ctx, query,
		shipment.ID, shipment.OrderID, shipment.UserID, shipment.ShippingCost, shipment.TrackingNumber, shipment.Carrier, shipment.Status, shipment.ShippingAddress, shipment.CreatedAt, shipment.UpdatedAt)
	if err != nil {
		// Handle unique constraint violation for tracking_number
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_shipments_tracking_number"` ||
			err.Error() == `pq: duplicate key value violates unique constraint "shipments_tracking_number_key"` {
			return errors.New("shipment with this tracking number already exists")
		}
		return fmt.Errorf("failed to save new shipment: %w", err)
	}
	return nil
}

// FindByID retrieves a shipment by its ID from PostgreSQL.
func (r *PostgreSQLShipmentRepository) FindByID(ctx context.Context, id string) (*domain.Shipment, error) {
	shipment := &domain.Shipment{}
	query := `SELECT id, order_id, user_id, shipping_cost, tracking_number, carrier, status, shipping_address, created_at, updated_at FROM shipments WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&shipment.ID, &shipment.OrderID, &shipment.UserID, &shipment.ShippingCost, &shipment.TrackingNumber,
		&shipment.Carrier, &shipment.Status, &shipment.ShippingAddress, &shipment.CreatedAt, &shipment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("shipment not found")
		}
		return nil, fmt.Errorf("failed to find shipment by ID: %w", err)
	}
	return shipment, nil
}

// FindByOrderID retrieves shipments associated with a specific order ID from PostgreSQL.
func (r *PostgreSQLShipmentRepository) FindByOrderID(ctx context.Context, orderID string) ([]*domain.Shipment, error) {
	var shipments []*domain.Shipment
	query := `SELECT id, order_id, user_id, shipping_cost, tracking_number, carrier, status, shipping_address, created_at, updated_at FROM shipments WHERE order_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query shipments by order ID: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		shipment := &domain.Shipment{}
		err := rows.Scan(&shipment.ID, &shipment.OrderID, &shipment.UserID, &shipment.ShippingCost, &shipment.TrackingNumber,
			&shipment.Carrier, &shipment.Status, &shipment.ShippingAddress, &shipment.CreatedAt, &shipment.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shipment row by order ID: %w", err)
		}
		shipments = append(shipments, shipment)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error for shipments by order ID: %w", err)
	}
	return shipments, nil
}

// FindByTrackingNumber retrieves a shipment by its tracking number from PostgreSQL.
func (r *PostgreSQLShipmentRepository) FindByTrackingNumber(ctx context.Context, trackingNumber string) (*domain.Shipment, error) {
	shipment := &domain.Shipment{}
	query := `SELECT id, order_id, user_id, shipping_cost, tracking_number, carrier, status, shipping_address, created_at, updated_at FROM shipments WHERE tracking_number = $1`
	row := r.db.QueryRowContext(ctx, query, trackingNumber)

	err := row.Scan(&shipment.ID, &shipment.OrderID, &shipment.UserID, &shipment.ShippingCost, &shipment.TrackingNumber,
		&shipment.Carrier, &shipment.Status, &shipment.ShippingAddress, &shipment.CreatedAt, &shipment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("shipment not found")
		}
		return nil, fmt.Errorf("failed to find shipment by tracking number: %w", err)
	}
	return shipment, nil
}

// FindAll retrieves a list of shipments based on filters and pagination from PostgreSQL.
func (r *PostgreSQLShipmentRepository) FindAll(ctx context.Context, userID, orderID, status string, limit, offset int32) ([]*domain.Shipment, int32, error) {
	var shipments []*domain.Shipment
	var totalCount int32

	baseQuery := `SELECT id, order_id, user_id, shipping_cost, tracking_number, carrier, status, shipping_address, created_at, updated_at FROM shipments`
	countQuery := `SELECT COUNT(*) FROM shipments`
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
		return nil, 0, fmt.Errorf("failed to get total shipment count: %w", err)
	}

	// Add pagination
	query := baseQuery + whereClause + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query shipments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		shipment := &domain.Shipment{}
		err := rows.Scan(&shipment.ID, &shipment.OrderID, &shipment.UserID, &shipment.ShippingCost, &shipment.TrackingNumber,
			&shipment.Carrier, &shipment.Status, &shipment.ShippingAddress, &shipment.CreatedAt, &shipment.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan shipment row: %w", err)
		}
		shipments = append(shipments, shipment)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return shipments, totalCount, nil
}

// Delete removes a shipment record from PostgreSQL by its ID.
func (r *PostgreSQLShipmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM shipments WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete shipment: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("shipment not found")
	}
	return nil
}
