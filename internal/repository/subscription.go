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

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

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

func (r *SubscriptionRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Subscription, error) {
	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
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

func (r *SubscriptionRepository) List(ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) ([]models.Subscription, int, error) {
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

func (r *SubscriptionRepository) GetCostSummary(ctx context.Context, userID *uuid.UUID, serviceName *string, startMonth, endMonth string) (int, error) {
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