package service

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionService_isValidDateFormat(t *testing.T) {
	service := &SubscriptionService{log: logrus.New()}

	tests := []struct {
		name     string
		date     string
		expected bool
	}{
		{"valid date", "01-2025", true},
		{"valid date december", "12-2025", true},
		{"invalid month", "13-2025", false},
		{"invalid month zero", "00-2025", false},
		{"invalid format", "1-2025", false},
		{"invalid format no dash", "012025", false},
		{"invalid year too low", "01-1899", false},
		{"invalid year too high", "01-2101", false},
		{"empty string", "", false},
		{"only dash", "-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidDateFormat(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscriptionService_isEndDateAfterStartDate(t *testing.T) {
	service := &SubscriptionService{log: logrus.New()}

	tests := []struct {
		name      string
		startDate string
		endDate   string
		expected  bool
	}{
		{"same date", "01-2025", "01-2025", true},
		{"end after start same year", "01-2025", "12-2025", true},
		{"end after start different year", "12-2024", "01-2025", true},
		{"end before start same year", "12-2025", "01-2025", false},
		{"end before start different year", "01-2025", "12-2024", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isEndDateAfterStartDate(tt.startDate, tt.endDate)
			assert.Equal(t, tt.expected, result)
		})
	}
}
