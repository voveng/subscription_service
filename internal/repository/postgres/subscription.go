package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"subscriptions-service/internal/model"

	"github.com/Masterminds/squirrel"
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
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.Insert("subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository.Create: failed to build query: %w", err)
	}

	var id uuid.UUID
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository.Create: %w", err)
	}
	return id, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	r.log.Info("repository: getting subscription by id", "id", id.String())
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repository.GetByID: failed to build query: %w", err)
	}

	sub := &model.Subscription{}
	err = r.db.QueryRow(ctx, query, args...).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository.GetByID: %w", err)
	}
	return sub, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repository.List: failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
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
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.Update("subscriptions").
		Set("service_name", sub.ServiceName).
		Set("price", sub.Price).
		Set("user_id", sub.UserID).
		Set("start_date", sub.StartDate).
		Set("end_date", sub.EndDate).
		Where(squirrel.Eq{"id": sub.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repository.Update: failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repository.Update: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.Delete("subscriptions").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repository.Delete: failed to build query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetSubscriptionsForTotalCost(ctx context.Context, userID uuid.UUID, serviceName, startDate, endDate string) ([]model.Subscription, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID})

	if serviceName != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"service_name": serviceName})
	}

	if startDate != "" {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"start_date": startDate})
	}

	if endDate != "" {
		queryBuilder = queryBuilder.Where(squirrel.Or{
			squirrel.Eq{"end_date": nil},
			squirrel.LtOrEq{"end_date": endDate},
		})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("repository.GetTotalCost: failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository.GetTotalCost: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate); err != nil {
			return nil, fmt.Errorf("repository.GetTotalCost: row scan failed: %w", err)
		}
		subs = append(subs, sub)
	}

	return subs, nil
}
