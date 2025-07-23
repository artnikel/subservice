package service

import (
	"regexp"
	"strconv"
	"strings"
)

// isValidDateFormat checks if a date string matches the MM-YYYY format
func (s *SubscriptionService) isValidDateFormat(date string) bool {
	pattern := `^(0[1-9]|1[0-2])-\d{4}$`
	matched, _ := regexp.MatchString(pattern, date)
	if !matched {
		return false
	}

	parts := strings.Split(date, "-")
	month, _ := strconv.Atoi(parts[0])
	year, _ := strconv.Atoi(parts[1])

	if month < 1 || month > 12 {
		return false
	}

	if year < 1900 || year > 2100 {
		return false
	}

	return true
}

// isEndDateAfterStartDate checks if endDate is lexicographically greater or equal to startDate
func (s *SubscriptionService) isEndDateAfterStartDate(startDate, endDate string) bool {
	start := strings.Replace(startDate, "-", "", 1)
	end := strings.Replace(endDate, "-", "", 1)

	startInt, _ := strconv.Atoi(start)
	endInt, _ := strconv.Atoi(end)

	return endInt >= startInt
}
