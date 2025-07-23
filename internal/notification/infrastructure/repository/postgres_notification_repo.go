// internal/notification/infrastructure/repository/postgres_notification_repo.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/notification/domain"
)

// PostgreSQLNotificationRepository implements the domain.NotificationRepository interface
// for PostgreSQL database operations.
type PostgreSQLNotificationRepository struct {
	db *sql.DB
}

// NewPostgreSQLNotificationRepository creates a new instance of PostgreSQLNotificationRepository.
func NewPostgreSQLNotificationRepository(db *sql.DB) *PostgreSQLNotificationRepository {
	return &PostgreSQLNotificationRepository{db: db}
}

// Save stores a notification record in PostgreSQL.
func (r *PostgreSQLNotificationRepository) Save(ctx context.Context, record *domain.NotificationRecord) error {
	// Check if record exists to decide between INSERT or UPDATE
	existingRecord, err := r.FindByID(ctx, record.ID)
	if err != nil && err.Error() != "notification record not found" {
		return fmt.Errorf("failed to check existing notification record: %w", err)
	}

	if existingRecord != nil {
		// Record exists, perform UPDATE
		query := `
            UPDATE notification_records
            SET user_id = $1, type = $2, recipient = $3, subject = $4, message = $5, status = $6, error_message = $7, updated_at = $8
            WHERE id = $9`
		_, err = r.db.ExecContext(ctx, query,
			record.UserID, record.Type, record.Recipient, record.Subject, record.Message, record.Status, record.Error, time.Now(), record.ID)
		if err != nil {
			return fmt.Errorf("failed to update notification record: %w", err)
		}
		return nil
	}

	// Record does not exist, perform INSERT
	query := `
        INSERT INTO notification_records (id, user_id, type, recipient, subject, message, status, error_message, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = r.db.ExecContext(ctx, query,
		record.ID, record.UserID, record.Type, record.Recipient, record.Subject, record.Message, record.Status, record.Error, record.SentAt, record.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save new notification record: %w", err)
	}
	return nil
}

// FindByID retrieves a notification record by its ID from PostgreSQL.
func (r *PostgreSQLNotificationRepository) FindByID(ctx context.Context, id string) (*domain.NotificationRecord, error) {
	record := &domain.NotificationRecord{}
	query := `SELECT id, user_id, type, recipient, subject, message, status, error_message, created_at, updated_at FROM notification_records WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var userID sql.NullString // Use NullString for optional UUID
	var subject sql.NullString
	var errorMessage sql.NullString

	err := row.Scan(&record.ID, &userID, &record.Type, &record.Recipient, &subject,
		&record.Message, &record.Status, &errorMessage, &record.SentAt, &record.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("notification record not found")
		}
		return nil, fmt.Errorf("failed to find notification record by ID: %w", err)
	}

	if userID.Valid {
		record.UserID = userID.String
	}
	if subject.Valid {
		record.Subject = subject.String
	}
	if errorMessage.Valid {
		record.Error = errorMessage.String
	}

	return record, nil
}

// FindAll retrieves a list of notification records based on filters from PostgreSQL.
func (r *PostgreSQLNotificationRepository) FindAll(ctx context.Context, userID, notificationType, status string, limit, offset int32) ([]*domain.NotificationRecord, int32, error) {
	var records []*domain.NotificationRecord
	var totalCount int32

	baseQuery := `SELECT id, user_id, type, recipient, subject, message, status, error_message, created_at, updated_at FROM notification_records`
	countQuery := `SELECT COUNT(*) FROM notification_records`
	args := []interface{}{}
	whereClause := ""
	argCounter := 1

	if userID != "" {
		whereClause += fmt.Sprintf(" WHERE user_id = $%d", argCounter)
		args = append(args, userID)
		argCounter++
	}
	if notificationType != "" {
		if whereClause == "" {
			whereClause += fmt.Sprintf(" WHERE type = $%d", argCounter)
		} else {
			whereClause += fmt.Sprintf(" AND type = $%d", argCounter)
		}
		args = append(args, notificationType)
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
		return nil, 0, fmt.Errorf("failed to get total notification record count: %w", err)
	}

	// Add pagination
	query := baseQuery + whereClause + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query notification records: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		record := &domain.NotificationRecord{}
		var userIDNull sql.NullString
		var subjectNull sql.NullString
		var errorMessageNull sql.NullString

		err := rows.Scan(&record.ID, &userIDNull, &record.Type, &record.Recipient, &subjectNull,
			&record.Message, &record.Status, &errorMessageNull, &record.SentAt, &record.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification record row: %w", err)
		}

		if userIDNull.Valid {
			record.UserID = userIDNull.String
		}
		if subjectNull.Valid {
			record.Subject = subjectNull.String
		}
		if errorMessageNull.Valid {
			record.Error = errorMessageNull.String
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return records, totalCount, nil
}
