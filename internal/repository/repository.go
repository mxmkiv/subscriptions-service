package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mxmkiv/subscriptions-service/internal/domain"
)

type ListFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
}

type SumFilter struct {
	UserID      uuid.UUID
	ServiceName *string
	StartDate   time.Time
	EndDate     time.Time
}

type Repository interface {
	Create(ctx context.Context, sub domain.CreateDTO) (domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, sub domain.UpdateDTO) (*domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) ([]domain.Subscription, error)
	SumByPeriod(ctx context.Context, filter SumFilter) (int, error)
}

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, dto domain.CreateDTO) (domain.Subscription, error) {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	sub := domain.Subscription{
		// id scan from db query
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserID:      dto.UserID,
		StartDate:   dto.StartDate,
		EndDate:     dto.EndDate,
	}

	err := r.db.QueryRowContext(ctx, query, dto.ServiceName, dto.Price, dto.UserID, dto.StartDate, dto.EndDate).Scan(&sub.ID)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query := `
		SELECT id, service_name, user_id, price, start_date, end_date
		FROM subscriptions
		WHERE id=$1
	`

	var sub domain.Subscription
	err := r.db.QueryRowContext(ctx, query, id).Scan(&sub.ID, &sub.ServiceName, &sub.UserID, &sub.Price, &sub.StartDate, &sub.EndDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateDTO) (*domain.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET service_name = $1, user_id = $2, price = $3, start_date = $4, end_date = $5
		WHERE id = $6
		RETURNING id, service_name, user_id, price, start_date, end_date
	`

	var sub domain.Subscription
	err := r.db.QueryRowContext(ctx, query, dto.ServiceName, dto.UserID, dto.Price, dto.StartDate, dto.EndDate, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.UserID, &sub.Price, &sub.StartDate, &sub.EndDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return &sub, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE from subscriptions where id=$1
	`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (r *postgresRepository) List(ctx context.Context, filter ListFilter) ([]domain.Subscription, error) {
	query := `
		SELECT id, service_name, user_id, price, start_date, end_date FROM subscriptions WHERE 1=1
	`
	var args []any
	argsIdx := 1

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argsIdx)
		args = append(args, *filter.ServiceName)
		argsIdx++
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argsIdx)
		args = append(args, *filter.UserID)
		//argsIdx++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.UserID, &sub.Price, &sub.StartDate, &sub.EndDate); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		subs = append(subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return subs, nil
}

func (r *postgresRepository) SumByPeriod(ctx context.Context, filter SumFilter) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions WHERE user_id=$1
		AND start_date <= $2
		AND (end_date IS NULL OR end_date >= $3)
	`

	args := []any{filter.UserID, filter.EndDate, filter.StartDate}
	argIdx := 4

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argIdx)
		args = append(args, *filter.ServiceName)
		//argIdx++
	}

	var total int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to sum subscriptions: %w", err)
	}

	return total, nil
}
