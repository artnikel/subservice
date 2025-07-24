package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	cerrors "github.com/artnikel/subservice/internal/errors"
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
				mockService.On("CreateSubscription", mock.Anything, mock.AnythingOfType("*models.CreateSubscriptionRequest")).Return(nil, cerrors.ErrInvalidStartDateFormat).Once()
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
				mockService.On("GetSubscription", mock.Anything, validID).Return(nil, cerrors.ErrSubscriptionNotFound).Once()
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

func TestSubscriptionHandler_UpdateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(mocks.MockSubscriptionService)
	logger := logrus.New()
	handler := NewSubscriptionHandler(mockService, logger)

	validID := uuid.New()

	serviceName := "Yandex Plus"
	price := 400
	startDate := "07-2025"
	endDate := "07-2026"

	updateReq := models.UpdateSubscriptionRequest{
		ServiceName: &serviceName,
		Price:       &price,
		StartDate:   &startDate,
		EndDate:     &endDate,
	}

	updatedSubscription := &models.Subscription{
		ID:          validID,
		ServiceName: serviceName,
		Price:       price,
		UserID:      uuid.New(),
		StartDate:   startDate,
		EndDate:     &endDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name           string
		id             string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "successful update",
			id:          validID.String(),
			requestBody: updateReq,
			mockSetup: func() {
				mockService.On("UpdateSubscription", mock.Anything, validID, mock.AnythingOfType("*models.UpdateSubscriptionRequest")).
					Return(updatedSubscription, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id format",
			id:             "invalid-uuid",
			requestBody:    updateReq,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			id:             validID.String(),
			requestBody:    "not json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "subscription not found",
			id:          validID.String(),
			requestBody: updateReq,
			mockSetup: func() {
				mockService.On("UpdateSubscription", mock.Anything, validID, mock.AnythingOfType("*models.UpdateSubscriptionRequest")).
					Return(nil, cerrors.ErrSubscriptionNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "service error",
			id:          validID.String(),
			requestBody: updateReq,
			mockSetup: func() {
				mockService.On("UpdateSubscription", mock.Anything, validID, mock.AnythingOfType("*models.UpdateSubscriptionRequest")).
					Return(nil, errors.New("db error")).Once()
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
			c.Request = httptest.NewRequest("PUT", "/subscriptions/"+tt.id, bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			handler.UpdateSubscription(c)

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
				mockService.On("DeleteSubscription", mock.Anything, id).Return(cerrors.ErrSubscriptionNotFound).Once()
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

func TestSubscriptionHandler_GetCostSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(mocks.MockSubscriptionService)
	logger := logrus.New()
	handler := NewSubscriptionHandler(mockService, logger)

	userID := uuid.NewString()
	serviceName := "Yandex Plus"

	req := models.CostSummaryRequest{
		UserID:      userID,
		ServiceName: &serviceName,
		StartMonth:  "01-2025",
		EndMonth:    "12-2025",
	}

	resp := &models.CostSummaryResponse{
		TotalCost: 12345,
	}

	params := url.Values{}
	params.Set("user_id", userID)
	params.Set("service_name", serviceName)
	params.Set("start_month", req.StartMonth)
	params.Set("end_month", req.EndMonth)

	queryString := "?" + params.Encode()

	tests := []struct {
		name           string
		queryString    string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "successful get cost summary",
			queryString: queryString,
			mockSetup: func() {
				mockService.On("GetCostSummary", mock.Anything, &req).Return(resp, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "missing required params",
			queryString: "?user_id=" + userID,
			mockSetup: func() {
				mockService.On("GetCostSummary", mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid user_id format",
			queryString: "?user_id=invalid-uuid&start_month=01-2025&end_month=12-2025",
			mockSetup: func() {
				mockService.On("GetCostSummary", mock.Anything, mock.Anything).Return(nil, errors.New("should not be called")).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			queryString: queryString,
			mockSetup: func() {
				mockService.On("GetCostSummary", mock.Anything, &req).Return(nil, errors.New("some error")).Once()
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
			c.Request = httptest.NewRequest("GET", "/subscriptions/cost_summary"+tt.queryString, http.NoBody)

			handler.GetCostSummary(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
