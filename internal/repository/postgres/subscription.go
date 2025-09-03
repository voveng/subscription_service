package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"subscriptions-service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type SubscriptionRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewSubscriptionRepository(db *pgxpool.Pool, log *slog.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, log: log}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) (uuid.UUID, error) {
	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository.Create: %w", err)
	}
	return id, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	r.log.Info("repository: getting subscription by id", "id", id.String())
	query := `SELECT id, service_name, price, user_id, start_date, end_date
              FROM subscriptions WHERE id = $1`
	sub := &model.Subscription{}
	err := r.db.QueryRow(ctx, query, id).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository.GetByID: %w", err)
	}
	return sub, nil
}

func (r *SubscriptionRepository) List(ctx context.Context) ([]model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository.List: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate); err != nil {
			return nil, fmt.Errorf("repository.List: row scan failed: %w", err)
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	query := `UPDATE subscriptions
              SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
              WHERE id = $6`
	_, err := r.db.Exec(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID)
	if err != nil {
		return fmt.Errorf("repository.Update: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName, startDate, endDate string) (int, error) {
	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE user_id = $1`
	args := []interface{}{userID}
	argID := 2

	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argID)
		args = append(args, serviceName)
		argID++
	}

	if startDate != "" {
		query += fmt.Sprintf(" AND start_date >= $%d", argID)
		args = append(args, startDate)
		argID++
	}

	if endDate != "" {
		query += fmt.Sprintf(" AND (end_date IS NULL OR end_date <= $%d)", argID)
		args = append(args, endDate)
	}

	var totalCost int
	err := r.db.QueryRow(ctx, query, args...).Scan(&totalCost)
	if err != nil {
		return 0, fmt.Errorf("repository.GetTotalCost: %w", err)
	}

	return totalCost, nil
}
