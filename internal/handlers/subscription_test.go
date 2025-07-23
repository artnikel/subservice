package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artnikel/subservice/internal/errors"
	"github.com/artnikel/subservice/internal/handlers/mocks"
	"github.com/artnikel/subservice/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscriptionHandler_CreateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(mocks.MockSubscriptionService)
	logger := logrus.New()
	handler := NewSubscriptionHandler(mockService, logger)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful creation",
			requestBody: models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "01-2025",
			},
			mockSetup: func() {
				subscription := &models.Subscription{
					ID:          uuid.New(),
					ServiceName: "Netflix",
					Price:       999,
					UserID:      uuid.New(),
					StartDate:   "01-2025",
				}
				mockService.On("CreateSubscription", mock.Anything, mock.AnythingOfType("*models.CreateSubscriptionRequest")).Return(subscription, nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: models.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   "invalid-date",
			},
			mockSetup: func() {
				mockService.On("CreateSubscription", mock.Anything, mock.AnythingOfType("*models.CreateSubscriptionRequest")).Return(nil, errors.ErrInvalidStartDateFormat).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/subscriptions", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateSubscription(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_GetSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(mocks.MockSubscriptionService)
	logger := logrus.New()
	handler := NewSubscriptionHandler(mockService, logger)

	validID := uuid.New()
	subscription := &models.Subscription{
		ID:          validID,
		ServiceName: "Netflix",
		Price:       999,
	}

	tests := []struct {
		name           string
		id             string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful get",
			id:   validID.String(),
			mockSetup: func() {
				mockService.On("GetSubscription", mock.Anything, validID).Return(subscription, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id format",
			id:             "invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "subscription not found",
			id:   validID.String(),
			mockSetup: func() {
				mockService.On("GetSubscription", mock.Anything, validID).Return(nil, errors.ErrSubscriptionNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", fmt.Sprintf("/subscriptions/%s", tt.id), http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			handler.GetSubscription(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_DeleteSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(mocks.MockSubscriptionService)
	logger := logrus.New()
	handler := NewSubscriptionHandler(mockService, logger)

	tests := []struct {
		name           string
		id             string
		mockSetup      func(uuid.UUID)
		expectedStatus int
	}{
		{
			name: "successful deletion",
			mockSetup: func(id uuid.UUID) {
				mockService.On("DeleteSubscription", mock.Anything, id).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid id format",
			id:             "invalid-uuid",
			mockSetup:      func(_ uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "subscription not found",
			mockSetup: func(id uuid.UUID) {
				mockService.On("DeleteSubscription", mock.Anything, id).Return(errors.ErrSubscriptionNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id uuid.UUID
			var idStr string

			if tt.id == "" {
				id = uuid.New()
				idStr = id.String()
			} else {
				idStr = tt.id
				var err error
				id, err = uuid.Parse(idStr)
				if err != nil {
					id = uuid.Nil
				}
			}

			tt.mockSetup(id)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("DELETE", "/subscriptions/"+idStr, http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: idStr}}

			handler.DeleteSubscription(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)

			mockService.ExpectedCalls = nil
			mockService.Calls = nil
		})
	}
}

func TestSubscriptionHandler_ListSubscriptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(mocks.MockSubscriptionService)
	logger := logrus.New()
	handler := NewSubscriptionHandler(mockService, logger)

	response := &models.ListResponse{
		Data:       []models.Subscription{},
		Total:      0,
		Page:       1,
		PageSize:   10,
		TotalPages: 0,
	}

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "successful list with defaults",
			queryParams: "",
			mockSetup: func() {
				mockService.On("ListSubscriptions", mock.Anything, (*uuid.UUID)(nil), (*string)(nil), 1, 10).Return(response, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "with query parameters",
			queryParams: "?page=2&page_size=20",
			mockSetup: func() {
				mockService.On("ListSubscriptions", mock.Anything, (*uuid.UUID)(nil), (*string)(nil), 2, 20).Return(response, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user_id format",
			queryParams:    "?user_id=invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/subscriptions"+tt.queryParams, http.NoBody)

			handler.ListSubscriptions(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
