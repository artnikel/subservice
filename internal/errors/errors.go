// Package errors defines common application-level errors used across services and handlers
package errors

import "errors"

var (
	// ErrSubscriptionNotFound is returned when a subscription with the given ID is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")

	// ErrInvalidStartDateFormat indicates that the provided start_date is not in the expected MM-YYYY format
	ErrInvalidStartDateFormat = errors.New("invalid start_date format, expected MM-YYYY")

	// ErrInvalidEndDateFormat indicates that the provided end_date is not in the expected MM-YYYY format
	ErrInvalidEndDateFormat = errors.New("invalid end_date format, expected MM-YYYY")

	// ErrEndDateBeforeStart is returned when the end_date is earlier than the start_date
	ErrEndDateBeforeStart = errors.New("end_date must be after start_date")

	// ErrInvalidStartMonthFormat indicates that the start_month is not in the expected MM-YYYY format
	ErrInvalidStartMonthFormat = errors.New("invalid start_month format, expected MM-YYYY")

	// ErrInvalidEndMonthFormat indicates that the end_month is not in the expected MM-YYYY format
	ErrInvalidEndMonthFormat = errors.New("invalid end_month format, expected MM-YYYY")

	// ErrEndMonthBeforeStart is returned when the end_month is earlier than the start_month
	ErrEndMonthBeforeStart = errors.New("end_month must be after or equal to start_month")
)
