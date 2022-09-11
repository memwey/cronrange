package cronrange

import (
	"time"
)

func (cr *CronRange) checkPrecondition() {
	switch {
	case cr == nil:
		panic("CronRange is nil")
	case cr.duration <= 0:
		panic("duration of CronRange is not positive")
	case cr.expr == nil:
		panic("schedule of CronRange is nil")
	}
}

// NextOccurrences returns the next occurrence time ranges, later than the given time.
//
// It panics if count is less than one, or the CronRange instance is nil or incomplete.
func (cr *CronRange) NextOccurrences(t time.Time, count int) (occurs []TimeRange) {
	cr.checkPrecondition()
	if count <= 0 {
		panic("count is not positive")
	}

	origLocation := t.Location()
	if cr.location != time.Local {
		t = t.In(cr.location)
	}

	for curr, i := t, 0; i < count; i++ {
		// if no occurrence is found within next five years, it returns zero time, i.e. time.Time{}
		next := cr.expr.Next(curr)
		if next.Before(curr) {
			break
		}
		occur := TimeRange{
			Start: next.In(origLocation),
			End:   next.Add(cr.duration).In(origLocation),
		}
		occurs = append(occurs, occur)
		curr = next
	}

	return
}

// IsWithin checks if the given time falls within any time range represented by the expression.
//
// It panics if the CronRange instance is nil or incomplete.
func (cr *CronRange) IsWithin(t time.Time) (within bool) {
	cr.checkPrecondition()

	if cr.location != time.Local {
		t = t.In(cr.location)
	}

	within = false
	searchStart := t.Add(-(cr.duration + 1*time.Second - 1*time.Nanosecond))
	rangeStart := cr.expr.Next(searchStart)
	rangeEnd := rangeStart.Add(cr.duration)

	// if no occurrence is found, it gets zero time, i.e. time.Time{}
	if rangeStart.Before(searchStart) {
		return
	}

	// check if rangeStart <= t <= rangeEnd
	within = (rangeStart.Before(t) && rangeEnd.After(t)) || rangeStart.Equal(t) || rangeEnd.Equal(t)
	return
}
