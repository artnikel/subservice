package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/artnikel/subservice/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) (*models.Subscription, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error  
	ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) (*models.ListResponse, error)
	GetCostSummary(ctx context.Context, req *models.CostSummaryRequest) (*models.CostSummaryResponse, error) 
}

type SubscriptionHandler struct {
	service SubscriptionService
	log     *logrus.Logger
}

func NewSubscriptionHandler(service SubscriptionService, log *logrus.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		log:     log,
	}
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Invalid request body for create subscription")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	subscription, err := h.service.CreateSubscription(c, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.log.WithError(err).Warn("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid subscription ID format"})
		return
	}

	subscription, err := h.service.GetSubscription(c, id)
	if err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})
			return
		}
		h.log.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.log.WithError(err).Warn("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid subscription ID format"})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("Invalid request body for update subscription")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	subscription, err := h.service.UpdateSubscription(c, id, &req)
	if err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})
			return
		}
		h.log.WithError(err).Error("Failed to update subscription")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.log.WithError(err).Warn("Invalid subscription ID format")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid subscription ID format"})
		return
	}

	err = h.service.DeleteSubscription(c, id)
	if err != nil {
		if err.Error() == "subscription not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})
			return
		}
		h.log.WithError(err).Error("Failed to delete subscription")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			h.log.WithError(err).Warn("Invalid user_id format")
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid user_id format"})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SubscriptionHandler) GetCostSummary(c *gin.Context) {
	var req models.CostSummaryRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.WithError(err).Warn("Invalid query parameters for cost summary")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := h.service.GetCostSummary(c, &req)
	if err != nil {
		h.log.WithError(err).Error("Failed to calculate cost summary")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
