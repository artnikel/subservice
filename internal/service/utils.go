package service

import (
	"fmt"
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
	const dateParts = 2
	parse := func(date string) (int, error) {
		parts := strings.Split(date, "-")
		if len(parts) != dateParts {
			return 0, fmt.Errorf("invalid date format: %s", date)
		}
		return strconv.Atoi(parts[1] + parts[0]) // YYYYMM
	}

	startInt, err1 := parse(startDate)
	endInt, err2 := parse(endDate)

	if err1 != nil || err2 != nil {
		return false
	}

	return endInt >= startInt
}
