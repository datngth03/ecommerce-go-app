// internal/recommendation/infrastructure/repository/postgres_user_interaction_repo.go
package repository

import (
	"context"
	"database/sql"

	// "errors"
	"fmt"
	// "time"

	"github.com/datngth03/ecommerce-go-app/internal/recommendation/domain" // Import domain package
)

// PostgreSQLUserInteractionRepository implements the domain.UserInteractionRepository interface.
// PostgreSQLUserInteractionRepository triển khai giao diện domain.UserInteractionRepository.
type PostgreSQLUserInteractionRepository struct {
	db *sql.DB
}

// NewPostgreSQLUserInteractionRepository creates a new instance of PostgreSQLUserInteractionRepository.
// NewPostgreSQLUserInteractionRepository tạo một thể hiện mới của PostgreSQLUserInteractionRepository.
func NewPostgreSQLUserInteractionRepository(db *sql.DB) *PostgreSQLUserInteractionRepository {
	return &PostgreSQLUserInteractionRepository{db: db}
}

// Save implements domain.UserInteractionRepository.
// Save triển khai phương thức Save của domain.UserInteractionRepository.
func (r *PostgreSQLUserInteractionRepository) Save(ctx context.Context, interaction *domain.UserInteraction) error {
	query := `
		INSERT INTO user_interactions (id, user_id, product_id, event_type, timestamp)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET user_id = $2, product_id = $3, event_type = $4, timestamp = $5;
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		interaction.ID,
		interaction.UserID,
		interaction.ProductID,
		interaction.EventType,
		interaction.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrFailedToSaveInteraction, err)
	}
	return nil
}

// FindByUserID implements domain.UserInteractionRepository.
// FindByUserID triển khai phương thức FindByUserID của domain.UserInteractionRepository.
func (r *PostgreSQLUserInteractionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.UserInteraction, error) {
	query := `
		SELECT id, user_id, product_id, event_type, timestamp
		FROM user_interactions
		WHERE user_id = $1
		ORDER BY timestamp DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
	}
	defer rows.Close()

	var interactions []*domain.UserInteraction
	for rows.Next() {
		var interaction domain.UserInteraction
		if err := rows.Scan(
			&interaction.ID,
			&interaction.UserID,
			&interaction.ProductID,
			&interaction.EventType,
			&interaction.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
		}
		interactions = append(interactions, &interaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
	}

	return interactions, nil
}

// FindPopularProducts implements domain.UserInteractionRepository.
// FindPopularProducts triển khai phương thức FindPopularProducts của domain.UserInteractionRepository.
func (r *PostgreSQLUserInteractionRepository) FindPopularProducts(ctx context.Context, limit int32) ([]*domain.UserInteraction, error) {
	// Simple approach: count interactions per product and order by count
	// Cách tiếp cận đơn giản: đếm tương tác cho mỗi sản phẩm và sắp xếp theo số lượng
	query := `
		SELECT product_id, COUNT(*) as interaction_count
		FROM user_interactions
		GROUP BY product_id
		ORDER BY interaction_count DESC
		LIMIT $1;
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
	}
	defer rows.Close()

	// Store product IDs with their interaction counts temporarily
	// Lưu tạm thời các ID sản phẩm với số lượng tương tác của chúng
	type ProductPopularity struct {
		ProductID string
		Count     int64
	}
	var popularProducts []ProductPopularity

	for rows.Next() {
		var pp ProductPopularity
		if err := rows.Scan(&pp.ProductID, &pp.Count); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
		}
		popularProducts = append(popularProducts, pp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
	}

	// For consistency with domain.UserInteraction, create dummy interactions
	// (In a real system, you might only return product IDs and fetch product details later)
	// Để nhất quán với domain.UserInteraction, tạo các tương tác giả
	// (Trong một hệ thống thực tế, bạn có thể chỉ trả về ID sản phẩm và lấy chi tiết sản phẩm sau)
	var interactions []*domain.UserInteraction
	for _, pp := range popularProducts {
		interactions = append(interactions, &domain.UserInteraction{
			ProductID: pp.ProductID,
			// Other fields are not relevant for "popular products" list based on this query
			// Các trường khác không liên quan đến danh sách "sản phẩm phổ biến" dựa trên truy vấn này
		})
	}

	return interactions, nil
}
