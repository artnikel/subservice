package service

import (
	"context"
	"testing"

	"github.com/artnikel/subservice/internal/errors"
	"github.com/artnikel/subservice/internal/models"
	"github.com/artnikel/subservice/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
				mockRepo.On("Delete", mock.Anything, id).Return(pgx.ErrNoRows).Once()
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

	userID := uuid.NewString()
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
				UserID:      userID,
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

func TestSubscriptionService_UpdateSubscription(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	logger := logrus.New()
	service := NewSubscriptionService(mockRepo, logger)

	id := uuid.New()
	existing := &models.Subscription{
		ID:        id,
		UserID:    uuid.New(),
		StartDate: "01-2025",
	}
	newName := "Spotify"
	newPrice := 1999
	newStart := "02-2025"
	newEnd := "12-2025"

	tests := []struct {
		name          string
		request       *models.UpdateSubscriptionRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful update",
			request: &models.UpdateSubscriptionRequest{
				ServiceName: &newName,
				Price:       &newPrice,
				StartDate:   &newStart,
				EndDate:     &newEnd,
			},
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, id).Return(existing, nil).Once()
				mockRepo.On("Update", mock.Anything, id, mock.AnythingOfType("*models.SubscriptionUpdates")).
					Return(&models.Subscription{ID: id, ServiceName: newName, Price: newPrice}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "invalid start date format",
			request: &models.UpdateSubscriptionRequest{
				StartDate: stringPtr("bad-format"),
			},
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, id).Return(existing, nil).Once()
			},
			expectedError: errors.ErrInvalidStartDateFormat,
		},
		{
			name: "end date before start date",
			request: &models.UpdateSubscriptionRequest{
				StartDate: stringPtr("02-2025"),
				EndDate:   stringPtr("01-2025"),
			},
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, id).Return(existing, nil).Once()
			},
			expectedError: errors.ErrEndDateBeforeStart,
		},
		{
			name: "subscription not found",
			request: &models.UpdateSubscriptionRequest{
				ServiceName: &newName,
			},
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, id).Return(nil, nil).Once()
			},
			expectedError: errors.ErrSubscriptionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.UpdateSubscription(context.Background(), id, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_ListSubscriptions(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	logger := logrus.New()
	service := NewSubscriptionService(mockRepo, logger)

	userID := uuid.New()
	serviceName := "Netflix"

	tests := []struct {
		name          string
		page          int
		pageSize      int
		mockSetup     func()
		expectedError error
		expectedTotal int
	}{
		{
			name:     "successful list",
			page:     1,
			pageSize: 10,
			mockSetup: func() {
				mockRepo.On("List", mock.Anything, &userID, &serviceName, 1, 10).
					Return([]models.Subscription{{ServiceName: "Netflix"}}, 1, nil).Once()
			},
			expectedError: nil,
			expectedTotal: 1,
		},
		{
			name:     "repository error",
			page:     1,
			pageSize: 10,
			mockSetup: func() {
				mockRepo.On("List", mock.Anything, &userID, &serviceName, 1, 10).
					Return(nil, 0, assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
		{
			name:     "invalid pagination values",
			page:     -5,
			pageSize: 200,
			mockSetup: func() {
				mockRepo.On("List", mock.Anything, &userID, &serviceName, 1, 10).
					Return([]models.Subscription{}, 0, nil).Once()
			},
			expectedError: nil,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.ListSubscriptions(context.Background(), &userID, &serviceName, tt.page, tt.pageSize)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTotal, result.Total)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
