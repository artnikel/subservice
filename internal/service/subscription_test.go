package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/artnikel/subservice/internal/errors"
	"github.com/artnikel/subservice/internal/models"
	"github.com/artnikel/subservice/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscriptionService_CreateSubscription(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	logger := logrus.New()
	service := NewSubscriptionService(mockRepo, logger)

	tests := []struct {
		name          string
		request       *models.CreateSubscriptionRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful creation",
			request: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "01-2025",
				EndDate:     nil,
			},
			mockSetup: func() {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Subscription")).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "invalid start date format",
			request: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "invalid-date",
				EndDate:     nil,
			},
			mockSetup:     func() {},
			expectedError: errors.ErrInvalidStartDateFormat,
		},
		{
			name: "end date before start date",
			request: &models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "12-2025",
				EndDate:     stringPtr("01-2025"),
			},
			mockSetup:     func() {},
			expectedError: errors.ErrEndDateBeforeStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.CreateSubscription(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.ServiceName, result.ServiceName)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_GetSubscription(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	logger := logrus.New()
	service := NewSubscriptionService(mockRepo, logger)

	id := uuid.New()
	subscription := &models.Subscription{
		ID:          id,
		ServiceName: "Netflix",
		Price:       999,
		UserID:      uuid.New(),
		StartDate:   "01-2025",
	}

	tests := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func()
		expectedError error
		expectedSub   *models.Subscription
	}{
		{
			name: "successful get",
			id:   id,
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, id).Return(subscription, nil).Once()
			},
			expectedError: nil,
			expectedSub:   subscription,
		},
		{
			name: "subscription not found",
			id:   id,
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, id).Return(nil, nil).Once()
			},
			expectedError: errors.ErrSubscriptionNotFound,
			expectedSub:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.GetSubscription(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSub, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_DeleteSubscription(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	logger := logrus.New()
	service := NewSubscriptionService(mockRepo, logger)

	id := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful deletion",
			id:   id,
			mockSetup: func() {
				mockRepo.On("Delete", mock.Anything, id).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "subscription not found",
			id:   id,
			mockSetup: func() {
				mockRepo.On("Delete", mock.Anything, id).Return(sql.ErrNoRows).Once()
			},
			expectedError: errors.ErrSubscriptionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			err := service.DeleteSubscription(context.Background(), tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_GetCostSummary(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	logger := logrus.New()
	service := NewSubscriptionService(mockRepo, logger)

	userID := uuid.New()
	serviceName := "Netflix"

	tests := []struct {
		name          string
		request       *models.CostSummaryRequest
		mockSetup     func()
		expectedError error
		expectedCost  int
	}{
		{
			name: "successful cost calculation",
			request: &models.CostSummaryRequest{
				UserID:      &userID,
				ServiceName: &serviceName,
				StartMonth:  "01-2025",
				EndMonth:    "12-2025",
			},
			mockSetup: func() {
				mockRepo.On("GetCostSummary", mock.Anything, &userID, &serviceName, "01-2025", "12-2025").Return(11988, nil).Once()
			},
			expectedError: nil,
			expectedCost:  11988,
		},
		{
			name: "invalid start month format",
			request: &models.CostSummaryRequest{
				StartMonth: "invalid",
				EndMonth:   "12-2025",
			},
			mockSetup:     func() {},
			expectedError: errors.ErrInvalidStartMonthFormat,
		},
		{
			name: "end month before start month",
			request: &models.CostSummaryRequest{
				StartMonth: "12-2025",
				EndMonth:   "01-2025",
			},
			mockSetup:     func() {},
			expectedError: errors.ErrEndMonthBeforeStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.GetCostSummary(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedCost, result.TotalCost)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
