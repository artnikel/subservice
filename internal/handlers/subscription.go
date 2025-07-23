// Package handlers provides HTTP handlers for managing subscriptions with validation and logging
package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	cerrors "github.com/artnikel/subservice/internal/errors"
	"github.com/artnikel/subservice/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SubscriptionService defines business logic methods used by the HTTP handlers
type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) (*models.Subscription, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) (*models.ListResponse, error)
	GetCostSummary(ctx context.Context, req *models.CostSummaryRequest) (*models.CostSummaryResponse, error)
}

// SubscriptionHandler handles HTTP requests related to subscriptions
type SubscriptionHandler struct {
	service SubscriptionService
	log     *logrus.Logger
}

// NewSubscriptionHandler creates a new SubscriptionHandler with service and logger dependencies
func NewSubscriptionHandler(service SubscriptionService, log *logrus.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		log:     log,
	}
}

// CreateSubscription handles POST requests to create a new subscription
// @Summary Create a subscription
// @Description Creates a new user subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} models.ErrorRequestResponse
// @Failure 500 {object} models.ErrorServerResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Invalid request body for create subscription")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: err.Error()})
		return
	}

	subscription, err := h.service.CreateSubscription(c, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription handles GET requests to fetch a subscription by ID
// @Summary Get Subscription
// @Description Gets subscriptions by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "subscription ID" format(uuid)
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorRequestResponse
// @Failure 404 {object} models.ErrorNotFoundResponse
// @Failure 500 {object} models.ErrorServerResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.log.WithError(err).Warn("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: "invalid subscription ID format"})
		return
	}

	subscription, err := h.service.GetSubscription(c, id)
	if err != nil {
		if errors.Is(err, cerrors.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorNotFoundResponse{Error: err.Error()})
			return
		}
		h.log.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, models.ErrorServerResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription handles PUT requests to update an existing subscription
// @Summary Update Subscription
// @Description Updates an existing subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID" format(uuid)
// @Param subscription body models.UpdateSubscriptionRequest true "Update data"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorRequestResponse
// @Failure 404 {object} models.ErrorNotFoundResponse
// @Failure 500 {object} models.ErrorServerResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.log.WithError(err).Warn("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: "invalid subscription ID format"})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Invalid request body for update subscription")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: err.Error()})
		return
	}

	subscription, err := h.service.UpdateSubscription(c, id, &req)
	if err != nil {
		if errors.Is(err, cerrors.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorNotFoundResponse{Error: err.Error()})
			return
		}
		h.log.WithError(err).Error("Failed to update subscription")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription handles DELETE requests to remove a subscription by ID
// @Summary Delete Subscription
// @Description Deletes subscription by ID
// @Tags subscriptions
// @Param id path string true "subscription ID" format(uuid)
// @Success 204
// @Failure 400 {object} models.ErrorRequestResponse
// @Failure 404 {object} models.ErrorNotFoundResponse
// @Failure 500 {object} models.ErrorServerResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.log.WithError(err).Warn("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: "invalid subscription ID format"})
		return
	}

	err = h.service.DeleteSubscription(c, id)
	if err != nil {
		if errors.Is(err, cerrors.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorNotFoundResponse{Error: err.Error()})
			return
		}
		h.log.WithError(err).Error("Failed to delete subscription")
		c.JSON(http.StatusInternalServerError, models.ErrorServerResponse{Error: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSubscriptions handles GET requests to list subscriptions with optional filters and pagination
// @Summary Get a list of subscriptions
// @Description Gets a list of subscriptions with pagination and filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID" format(uuid)
// @Param service_name query string false "Service name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.ListResponse
// @Failure 400 {object} models.ErrorRequestResponse
// @Failure 500 {object} models.ErrorServerResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			h.log.WithError(err).Warn("Invalid user_id format")
			c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: "invalid user_id format"})
			return
		}
		userID = &id
	}

	var serviceName *string
	if serviceNameStr := c.Query("service_name"); serviceNameStr != "" {
		serviceName = &serviceNameStr
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 10
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	response, err := h.service.ListSubscriptions(c, userID, serviceName, page, pageSize)
	if err != nil {
		h.log.WithError(err).Error("Failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, models.ErrorServerResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCostSummary handles GET requests to calculate total subscription cost for a period with filters
// @Summary Calculate total cost
// @Description Calculates the total cost of subscriptions per period with filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID" format(uuid)
// @Param service_name query string false "Service name"
// @Param start_month query string true "Start month" format(MM- YYYY)
// @Param end_month query string true "End month" format(MM-YYYY)
// @Success 200 {object} models.CostSummaryResponse
// @Failure 400 {object} models.ErrorRequestResponse
// @Failure 500 {object} models.ErrorServerResponse
// @Router /cost/summary [get]
func (h *SubscriptionHandler) GetCostSummary(c *gin.Context) {
	var req models.CostSummaryRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.WithError(err).Warn("Invalid query parameters for cost summary")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: err.Error()})
		return
	}

	response, err := h.service.GetCostSummary(c, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to calculate cost summary")
		c.JSON(http.StatusBadRequest, models.ErrorRequestResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
