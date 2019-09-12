package scheduleutil

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"

	"github.com/dsbezerra/amenic/src/lib/util/timeutil"
)

type (
	Period struct {
		Start time.Time `json:"start" bson:"start,omitempty"`
		End   time.Time `json:"end" bson:"end,omitempty"`
	}
)

// GetWeekPeriodID returns the ID of the current period of movie screening for the given time
func GetWeekPeriodID(now *time.Time) string {
	t := GetWeekPeriod(now)

	// Get start and end as YYYY-MM-DD strings.
	s := timeutil.TimeToSimpleDateString(&t.Start)
	e := timeutil.TimeToSimpleDateString(&t.End)

	if s != "" && e != "" {
		// Format to YYYY-MM-DD,YYYY-MM-DD.
		str := fmt.Sprintf("%s,%s", s, e)

		// Get the hash.
		h := md5.New()
		io.WriteString(h, str)
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	return ""
}

// GetWeekPeriod returns current period of movie screening for the given time
func GetWeekPeriod(t *time.Time) *Period {
	if t == nil {
		loc, _ := time.LoadLocation("America/Sao_Paulo")
		now := time.Now().In(loc)
		t = &now
	}

	count := DaysUntilNextWednesday(t)
	y, m, d, loc := t.Year(), t.Month(), t.Day(), t.Location()

	s := time.Date(y, m, d, 0, 0, 0, 0, loc)
	s = s.AddDate(0, 0, -(6 - count))

	e := time.Date(y, m, d, 0, 0, 0, 0, loc)
	e = e.AddDate(0, 0, count)

	return &Period{Start: s.UTC(), End: e.UTC()}
}

// DaysUntilNextWednesday calculates how many days we are to next wednesday
func DaysUntilNextWednesday(now *time.Time) int {
	result := -1

	w := int(now.Weekday())
	if w < 4 {
		result = (4 - 1) - w
	} else {
		result = (4 + 7 - 1) - w
	}

	return result
}
