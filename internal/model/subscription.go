package model

import "github.com/google/uuid"

// Subscription represents a user's subscription to a service.
// @Description Subscription information
type Subscription struct {
	ID           uuid.UUID `json:"id,omitempty"`
	ServiceName  string    `json:"service_name" binding:"required"`
	Price        int       `json:"price" binding:"required,gte=0"`
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	StartDate    string    `json:"start_date" binding:"required"` // Format: MM-YYYY
	EndDate      *string   `json:"end_date,omitempty"`           // Format: MM-YYYY
}

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" binding:"required"`
	Price       int       `json:"price" binding:"required,gte=0"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	StartDate   string    `json:"start_date" binding:"required"` // Format: MM-YYYY
	EndDate     *string   `json:"end_date,omitempty"`           // Format: MM-YYYY
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty,gte=0"`
	StartDate   *string `json:"start_date,omitempty"` // Format: MM-YYYY
	EndDate     *string `json:"end_date,omitempty"`   // Format: MM-YYYY
}