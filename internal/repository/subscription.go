// Package repository provides data access layer for subscriptions in PostgreSQL using pgxpool
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/artnikel/subservice/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubscriptionRepository manages subscription persistence in the database
type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

// NewSubscriptionRepository creates a new repository with the given pgx connection pool
func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

// Create inserts a new subscription into the database and updates the model with generated fields
func (r *SubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx,
		query,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserID,
		subscription.StartDate,
		subscription.EndDate,
	).Scan(&subscription.ID, &subscription.CreatedAt, &subscription.UpdatedAt)
}

// GetByID fetches a subscription by its ID or returns nil if not found
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	subscription := &models.Subscription{}
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.ServiceName,
		&subscription.Price,
		&subscription.UserID,
		&subscription.StartDate,
		&subscription.EndDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return subscription, err
}

// Update modifies fields of a subscription specified in updates map and returns the updated subscription
func (r *SubscriptionRepository) Update(ctx context.Context, id uuid.UUID, updates *models.SubscriptionUpdates) (*models.Subscription, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.ServiceName != nil {
		setParts = append(setParts, fmt.Sprintf("service_name = $%d", argIndex))
		args = append(args, *updates.ServiceName)
		argIndex++
	}

	if updates.Price != nil {
		setParts = append(setParts, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, *updates.Price)
		argIndex++
	}

	if updates.StartDate != nil {
		setParts = append(setParts, fmt.Sprintf("start_date = $%d", argIndex))
		args = append(args, *updates.StartDate)
		argIndex++
	}

	if updates.EndDate != nil {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIndex))
		args = append(args, *updates.EndDate)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE subscriptions 
		SET %s
		WHERE id = $%d
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	subscription := &models.Subscription{}
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&subscription.ID,
		&subscription.ServiceName,
		&subscription.Price,
		&subscription.UserID,
		&subscription.StartDate,
		&subscription.EndDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return subscription, err
}

// Delete removes a subscription by ID, returning sql.ErrNoRows if none deleted
func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List retrieves subscriptions filtered by optional userID and serviceName with pagination, returning total count
func (r *SubscriptionRepository) List(
	ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) ([]models.Subscription, int, error) {
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
	}

	if serviceName != nil && *serviceName != "" {
		conditions = append(conditions, fmt.Sprintf("service_name ILIKE $%d", argIndex))
		args = append(args, "%"+*serviceName+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM subscriptions %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	listQuery := fmt.Sprintf(`
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subscriptions := []models.Subscription{}
	for rows.Next() {
		var s models.Subscription
		err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		subscriptions = append(subscriptions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return subscriptions, total, nil
}

// GetCostSummary calculates total subscription cost filtered by optional userID, serviceName and date range
func (r *SubscriptionRepository) GetCostSummary(
	ctx context.Context, userID *uuid.UUID, serviceName *string, startMonth, endMonth string) (int, error) {
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
	}

	if serviceName != nil && *serviceName != "" {
		conditions = append(conditions, fmt.Sprintf("service_name ILIKE $%d", argIndex))
		args = append(args, "%"+*serviceName+"%")
		argIndex++
	}

	periodCondition := fmt.Sprintf(`
		(start_date <= $%d AND (end_date IS NULL OR end_date >= $%d))`,
		argIndex, argIndex+1)
	conditions = append(conditions, periodCondition)
	args = append(args, endMonth, startMonth)

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(price), 0) as total_cost
		FROM subscriptions %s`, whereClause)

	var totalCost int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&totalCost)
	return totalCost, err
}
