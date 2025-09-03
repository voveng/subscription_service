package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"subscriptions-service/internal/model"
	"subscriptions-service/internal/repository/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubscriptionService interface {
	Create(ctx context.Context, sub *model.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context) ([]model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName, startDate, endDate string) (int, error)
}

type Handler struct {
	service SubscriptionService
	log     *slog.Logger
}

func NewHandler(service SubscriptionService, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// Create godoc
// @Summary      Create a subscription
// @Description  Create a new subscription
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        input body model.Subscription true "Subscription Info"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions [post]
func (h *Handler) Create(c *gin.Context) {
	var sub model.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		h.log.Error("failed to bind json", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &sub)
	if err != nil {
		h.log.Error("failed to create subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetByID godoc
// @Summary      Get a subscription by ID
// @Description  Get a single subscription by its ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	h.log.Info("handler: attempting to get subscription by id", "id", id.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// List godoc
// @Summary      List subscriptions
// @Description  Get a list of all subscriptions
// @Tags         subscriptions
// @Produce      json
// @Success      200  {array}   model.Subscription
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions [get]
func (h *Handler) List(c *gin.Context) {
	subs, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// Update godoc
// @Summary      Update a subscription
// @Description  Update an existing subscription
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Param        input body model.Subscription true "Subscription Info"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var sub model.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub.ID = id
	if err := h.service.Update(c.Request.Context(), &sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Delete godoc
// @Summary      Delete a subscription
// @Description  Delete a subscription by its ID
// @Tags         subscriptions
// @Param        id   path      string  true  "Subscription ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTotalCost godoc
// @Summary      Get total cost of subscriptions
// @Description  Get total cost of subscriptions for a user, with optional filters
// @Tags         subscriptions
// @Produce      json
// @Param        user_id      query     string  true  "User ID"
// @Param        service_name query     string  false "Service Name"
// @Param        start_date   query     string  false "Start Date (MM-YYYY)"
// @Param        end_date     query     string  false "End Date (MM-YYYY)"
// @Success      200  {object}  map[string]int
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/total_cost [get]
func (h *Handler) GetTotalCost(c *gin.Context) {
	userID, err := uuid.Parse(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	serviceName := c.Query("service_name")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	totalCost, err := h.service.GetTotalCost(c.Request.Context(), userID, serviceName, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get total cost"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_cost": totalCost})
}
