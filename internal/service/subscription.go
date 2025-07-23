package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/artnikel/subservice/internal/models"
	"github.com/artnikel/subservice/internal/repository"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
	log  *logrus.Logger
}

func NewSubscriptionService(repo *repository.SubscriptionRepository, log *logrus.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
		log:  log,
	}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	s.log.WithFields(logrus.Fields{
		"service_name": req.ServiceName,
		"user_id":      req.UserID,
		"price":        req.Price,
	}).Info("Creating new subscription")

	if !s.isValidDateFormat(req.StartDate) {
		return nil, errors.New("invalid start_date format, expected MM-YYYY")
	}

	if req.EndDate != nil && !s.isValidDateFormat(*req.EndDate) {
		return nil, errors.New("invalid end_date format, expected MM-YYYY")
	}

	if req.EndDate != nil && !s.isEndDateAfterStartDate(req.StartDate, *req.EndDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	subscription := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := s.repo.Create(ctx, subscription); err != nil {
		s.log.WithError(err).Error("Failed to create subscription")
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.log.WithField("subscription_id", subscription.ID).Info("Subscription created successfully")
	return subscription, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	s.log.WithField("subscription_id", id).Info("Getting subscription")

	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get subscription")
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		s.log.WithField("subscription_id", id).Warn("Subscription not found")
		return nil, errors.New("subscription not found")
	}

	return subscription, nil
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	s.log.WithField("subscription_id", id).Info("Updating subscription")

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("Failed to get subscription for update")
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if existing == nil {
		s.log.WithField("subscription_id", id).Warn("Subscription not found for update")
		return nil, errors.New("subscription not found")
	}

	updates := make(map[string]interface{})

	if req.ServiceName != nil {
		updates["service_name"] = *req.ServiceName
	}

	if req.Price != nil {
		updates["price"] = *req.Price
	}

	if req.StartDate != nil {
		if !s.isValidDateFormat(*req.StartDate) {
			return nil, errors.New("invalid start_date format, expected MM-YYYY")
		}
		updates["start_date"] = *req.StartDate
	}

	if req.EndDate != nil {
		if !s.isValidDateFormat(*req.EndDate) {
			return nil, errors.New("invalid end_date format, expected MM-YYYY")
		}
		updates["end_date"] = *req.EndDate
	}

	startDate := existing.StartDate
	if req.StartDate != nil {
		startDate = *req.StartDate
	}

	if req.EndDate != nil && !s.isEndDateAfterStartDate(startDate, *req.EndDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	subscription, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		s.log.WithError(err).Error("Failed to update subscription")
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	s.log.WithField("subscription_id", id).Info("Subscription updated successfully")
	return subscription, nil
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	s.log.WithField("subscription_id", id).Info("Deleting subscription")

	err := s.repo.Delete(ctx, id)
	if err == sql.ErrNoRows {
		s.log.WithField("subscription_id", id).Warn("Subscription not found for deletion")
		return errors.New("subscription not found")
	}

	if err != nil {
		s.log.WithError(err).Error("Failed to delete subscription")
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.log.WithField("subscription_id", id).Info("Subscription deleted successfully")
	return nil
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) (*models.ListResponse, error) {
	s.log.WithFields(logrus.Fields{
		"user_id":      userID,
		"service_name": serviceName,
		"page":         page,
		"page_size":    pageSize,
	}).Info("Listing subscriptions")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	subscriptions, total, err := s.repo.List(ctx, userID, serviceName, page, pageSize)
	if err != nil {
		s.log.WithError(err).Error("Failed to list subscriptions")
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	totalPages := (total + pageSize - 1) / pageSize

	response := &models.ListResponse{
		Data:       subscriptions,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	s.log.WithFields(logrus.Fields{
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
	}).Info("Subscriptions listed successfully")

	return response, nil
}

func (s *SubscriptionService) GetCostSummary(ctx context.Context, req *models.CostSummaryRequest) (*models.CostSummaryResponse, error) {
	s.log.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"service_name": req.ServiceName,
		"start_month":  req.StartMonth,
		"end_month":    req.EndMonth,
	}).Info("Calculating cost summary")

	if !s.isValidDateFormat(req.StartMonth) {
		return nil, errors.New("invalid start_month format, expected MM-YYYY")
	}

	if !s.isValidDateFormat(req.EndMonth) {
		return nil, errors.New("invalid end_month format, expected MM-YYYY")
	}

	if !s.isEndDateAfterStartDate(req.StartMonth, req.EndMonth) {
		return nil, errors.New("end_month must be after or equal to start_month")
	}

	totalCost, err := s.repo.GetCostSummary(ctx, req.UserID, req.ServiceName, req.StartMonth, req.EndMonth)
	if err != nil {
		s.log.WithError(err).Error("Failed to calculate cost summary")
		return nil, fmt.Errorf("failed to calculate cost summary: %w", err)
	}

	response := &models.CostSummaryResponse{
		TotalCost: totalCost,
	}

	s.log.WithField("total_cost", totalCost).Info("Cost summary calculated successfully")
	return response, nil
}

func (s *SubscriptionService) isValidDateFormat(date string) bool {
	pattern := `^\d{2}-\d{4}$`
	matched, _ := regexp.MatchString(pattern, date)
	return matched
}

func (s *SubscriptionService) isEndDateAfterStartDate(startDate, endDate string) bool {
	return endDate >= startDate
}
