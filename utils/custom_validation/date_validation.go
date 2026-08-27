package customvalidation

import (
	"errors"
	"fmt"
	"time"
)

var MaxRangeDays = 60
const LocalTimeLayout = "2006-01-02T15:04:05"

var (
	ErrStartDateRequired = errors.New("start date is required")
	ErrEndDateRequired   = errors.New("end date is required")
	ErrStartDateInPast   = errors.New("start date must be in the future")
	ErrEndDateInPast   = errors.New("end date must be in the future")
	ErrEndBeforeStart    = errors.New("end date cannot be before start date")
	ErrRangeTooLong      = fmt.Errorf("date range cannot be longer than %d days", MaxRangeDays)
	ErrInvalidTimeZone = errors.New("invalid time zone, use an IANA name like Asia/Dhaka")
)

func ValidateDateRange(startDate, endDate time.Time) (time.Time, time.Time, error) {

	if startDate.IsZero() {
		return time.Time{}, time.Time{}, ErrStartDateRequired
	}

	if endDate.IsZero() {
		return time.Time{}, time.Time{}, ErrEndDateRequired
	}

	today := time.Now()

	startDate, _ = ParseLocalTime(startDate.Format(time.RFC3339), "Local")
	endDate, _ = ParseLocalTime(endDate.Format(time.RFC3339), "Local")
	today, _ = ParseLocalTime(today.Format(time.RFC3339), "Local")


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

	// if dateRange == 0 {
	// 	formatedEndDate = formatedEndDate.AddDate(0, 0, 1)
	// }

	return formatedStartDate, formatedEndDate, nil
}


func ParseLocalTime(value, timeZoneName string) (time.Time, error) {
	timeZoneName = "Local"
	loc, err := time.LoadLocation(timeZoneName)
	if err != nil {
		return time.Time{}, ErrInvalidTimeZone
	}

	t, err := time.ParseInLocation(LocalTimeLayout, value, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time %q, expected format %s", value, LocalTimeLayout)
	}

	return t.UTC(), nil

}