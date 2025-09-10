package service

import (
	"context"
	"log/slog"
	"subscriptions-service/internal/model"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -source=subscription.go -destination=mocks/mock.go
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, limit, offset int) ([]model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetSubscriptionsForTotalCost(ctx context.Context, userID uuid.UUID, serviceName, startDate, endDate string) ([]model.Subscription, error)
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
	const op = "service.GetByID"
	log := s.log.With(slog.String("op", op))

	log.Info("getting subscription by id", "id", id.String())
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("failed to get subscription by id", "error", err)
		return nil, err
	}
	log.Info("got subscription by id successfully", "id", id.String())
	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	const op = "service.List"
	log := s.log.With(slog.String("op", op))

	log.Info("listing subscriptions")
	subs, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		log.Error("failed to list subscriptions", "error", err)
		return nil, err
	}
	log.Info("listed subscriptions successfully", "count", len(subs))
	return subs, nil
}

func (s *SubscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	const op = "service.Update"
	log := s.log.With(slog.String("op", op))

	log.Info("updating subscription", "id", sub.ID.String())

	_, err := s.repo.GetByID(ctx, sub.ID)
	if err != nil {
		log.Error("failed to get subscription before update", "error", err)
		return err
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		log.Error("failed to update subscription", "error", err)
		return err
	}
	log.Info("updated subscription successfully", "id", sub.ID.String())
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "service.Delete"
	log := s.log.With(slog.String("op", op))

	log.Info("deleting subscription", "id", id.String())

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("failed to get subscription before delete", "error", err)
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error("failed to delete subscription", "error", err)
		return err
	}
	log.Info("deleted subscription successfully", "id", id.String())
	return nil
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName, startDate, endDate string) (int, error) {
	const op = "service.GetTotalCost"
	log := s.log.With(slog.String("op", op))

	log.Info("getting total cost")
	subs, err := s.repo.GetSubscriptionsForTotalCost(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		log.Error("failed to get subscriptions for total cost", "error", err)
		return 0, err
	}

	var totalCost int
	monthlyCosts := make(map[time.Month]map[int]int)

	for _, sub := range subs {
		start, err := time.Parse("2006-01-02", sub.StartDate)
		if err != nil {
			log.Error("failed to parse start date", "error", err)
			continue
		}

		end := time.Now().AddDate(10, 0, 0) // 10 years in the future for open-ended subscriptions
		if sub.EndDate != nil {
			end, err = time.Parse("2006-01-02", *sub.EndDate)
			if err != nil {
				log.Error("failed to parse end date", "error", err)
				continue
			}
		}

		for d := start; d.Before(end); d = d.AddDate(0, 1, 0) {
			if monthlyCosts[d.Month()] == nil {
				monthlyCosts[d.Month()] = make(map[int]int)
			}
			monthlyCosts[d.Month()][d.Year()] += sub.Price
		}
	}

	for _, yearCosts := range monthlyCosts {
		for _, cost := range yearCosts {
			totalCost += cost
		}
	}

	log.Info("got total cost successfully", "total_cost", totalCost)
	return totalCost, nil
}
