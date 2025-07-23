package models

import (
	"time"

	"github.com/google/uuid"
)

// Subscription introduces the subscription model
type Subscription struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	Price       int       `json:"price" db:"price"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	StartDate   string    `json:"start_date" db:"start_date"`       // MM-YYYY
	EndDate     *string   `json:"end_date,omitempty" db:"end_date"` // MM-YYYY
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateSubscriptionRequest represents a request to create a subscription
type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int       `json:"price" binding:"required,min=1" example:"400"`
	UserID      uuid.UUID `json:"user_id" binding:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"07-2026"`
}

// UpdateSubscriptionRequest represents a request to update a subscription
type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus"`
	Price       *int    `json:"price,omitempty" binding:"omitempty,min=1" example:"400"`
	StartDate   *string `json:"start_date,omitempty" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"07-2026"`
}

// CostSummaryRequest represents a request to calculate a cost
type CostSummaryRequest struct {
	UserID      *uuid.UUID `form:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName *string    `form:"service_name" example:"Yandex Plus"`
	StartMonth  string     `form:"start_month" binding:"required" example:"01-2025"`
	EndMonth    string     `form:"end_month" binding:"required" example:"12-2025"`
}

// CostSummaryResponse represents a response with a summarized cost
type CostSummaryResponse struct {
	TotalCost int `json:"total_cost" example:"4800"`
}

// ErrorRequestResponse represents an requset error response 
type ErrorRequestResponse struct {
	Error string `json:"error" example:"Invalid request"`
}

// ErrorNotFoundResponse represents an not found error response
type ErrorNotFoundResponse struct {
	Error string `json:"error" example:"Not found"`
}

// ErrorServerResponse represents an server error response
type ErrorServerResponse struct {
	Error string `json:"error" example:"Internal server error"`
}

// ListResponse presents a response with a list of subscriptions
type ListResponse struct {
	Data       []Subscription `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
