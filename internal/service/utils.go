package service

import "regexp"

// isValidDateFormat checks if a date string matches the MM-YYYY format
func (s *SubscriptionService) isValidDateFormat(date string) bool {
	pattern := `^\d{2}-\d{4}$`
	matched, _ := regexp.MatchString(pattern, date)
	return matched
}

// isEndDateAfterStartDate checks if endDate is lexicographically greater or equal to startDate
func (s *SubscriptionService) isEndDateAfterStartDate(startDate, endDate string) bool {
	return endDate >= startDate
}
