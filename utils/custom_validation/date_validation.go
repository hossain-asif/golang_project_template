package customvalidation

import (
	"errors"
	"fmt"
	"time"
)

var MaxRangeDays = 60

var (
	ErrStartDateRequired = errors.New("start date is required")
	ErrEndDateRequired   = errors.New("end date is required")
	ErrStartDateInPast   = errors.New("start date cannot be in the past")
	ErrEndDateInPast   = errors.New("end date cannot be in the past")
	ErrEndBeforeStart    = errors.New("end date cannot be before start date")
	ErrRangeTooLong      = fmt.Errorf("date range cannot be longer than %d days", MaxRangeDays)
)

func ValidateDateRange(startDate, endDate time.Time) (time.Time, time.Time, error) {

	if startDate.IsZero() {
		return time.Time{}, time.Time{}, ErrStartDateRequired
	}

	if endDate.IsZero() {
		return time.Time{}, time.Time{}, ErrEndDateRequired
	}

	startDateYear, startDateMonth, startDateDay := startDate.Date()
	formatedStartDate := time.Date(startDateYear, startDateMonth, startDateDay, 0, 0, 0, 0, time.UTC)

	endDateYear, endDateMonth, endDateDay := endDate.Date()
	formatedEndDate := time.Date(endDateYear, endDateMonth, endDateDay, 0, 0, 0, 0, time.UTC)

	todayYear, todayMonth, todayDay := time.Now().Date()
	formatedToday := time.Date(todayYear, todayMonth, todayDay, 0, 0, 0, 0, time.UTC)


	if formatedStartDate.Before(formatedToday) {
		return time.Time{}, time.Time{}, ErrStartDateInPast
	}

	if formatedEndDate.Before(formatedToday) {
		return time.Time{}, time.Time{}, ErrEndDateInPast
	}

	if formatedEndDate.Before(formatedStartDate) {
		return time.Time{}, time.Time{}, ErrEndBeforeStart
	}

	dateRange := int(formatedEndDate.Sub(formatedStartDate).Hours() / 24)
	if dateRange > MaxRangeDays {
		return time.Time{}, time.Time{}, ErrRangeTooLong
	}

	if dateRange == 0 {
		formatedEndDate = formatedEndDate.AddDate(0, 0, 1)
	}

	return formatedStartDate, formatedEndDate, nil
}