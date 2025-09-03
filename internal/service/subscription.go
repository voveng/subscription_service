package service

import (
	"context"
	"log/slog"
	"subscriptions-service/internal/model"

	"github.com/google/uuid"
)

//go:generate mockgen -source=subscription.go -destination=mocks/mock.go
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context) ([]model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName, startDate, endDate string) (int, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
	log  *slog.Logger
}

func NewSubscriptionService(repo SubscriptionRepository, log *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, log: log}
}

func (s *SubscriptionService) Create(ctx context.Context, sub *model.Subscription) (uuid.UUID, error) {
	const op = "service.Create"
	log := s.log.With(slog.String("op", op))

	log.Info("creating subscription")
	id, err := s.repo.Create(ctx, sub)
	if err != nil {
		log.Error("failed to create subscription", "error", err)
		return uuid.Nil, err
	}
	log.Info("subscription created successfully", "id", id)
	return id, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	return s.repo.List(ctx)
}

func (s *SubscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	return s.repo.Update(ctx, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName, startDate, endDate string) (int, error) {
	return s.repo.GetTotalCost(ctx, userID, serviceName, startDate, endDate)
}
