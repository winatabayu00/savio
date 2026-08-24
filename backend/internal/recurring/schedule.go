// Package recurring generates scheduled occurrence dates for recurring
// rules. All generators are pure functions over a start date so the schedule
// is deterministic (testable without a database).
package recurring

import "time"

// nextDates returns the occurrence due dates from day (the day the schedule
// begins, inclusive) up to and including horizon. Dates are truncated to UTC
// midnight before comparison.
func nextDates(freq string, start time.Time, end *time.Time, horizon time.Time) []time.Time {
	start = atDate(start)
	horizon = atDate(horizon)
	if horizon.Before(start) {
		return nil
	}
	limit := horizon
	if end != nil {
		e := atDate(*end)
		if e.Before(limit) {
			limit = e
		}
	}
	var out []time.Time
	switch Frequency(freq) {
	case FreqDaily:
		for d := start; !d.After(limit); d = d.AddDate(0, 0, 1) {
			out = append(out, d)
		}
	case FreqWeekly:
		for d := start; !d.After(limit); d = d.AddDate(0, 0, 7) {
			out = append(out, d)
		}
	case FreqMonthly:
		dom := start.Day()
		if !start.After(limit) {
			out = append(out, start)
		}
		// step on month-day 1 so AddDate never overflows
		m := firstDayOfMonth(start).AddDate(0, 1, 0)
		for {
			cand := clampDay(m, dom)
			if cand.After(limit) {
				break
			}
			out = append(out, cand)
			m = m.AddDate(0, 1, 0)
		}
	case FreqMonthEnd:
		// last day of each month, starting from start's own month
		m := firstDayOfMonth(start)
		for {
			cand := lastDayOfMonth(m)
			if cand.After(limit) {
				break
			}
			if !cand.Before(start) {
				out = append(out, cand)
			}
			m = m.AddDate(0, 1, 0)
		}
	}
	return out
}

func firstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// clampDay returns the date of month m with day-of-month dom, clamped to the
// month's last day when the month is shorter (e.g. day 31 → Feb 28/29).
func clampDay(m time.Time, dom int) time.Time {
	last := lastDayOfMonth(m)
	if dom > last.Day() {
		return last
	}
	return time.Date(m.Year(), m.Month(), dom, 0, 0, 0, 0, time.UTC)
}

func atDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func lastDayOfMonth(t time.Time) time.Time {
	first := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	next := first.AddDate(0, 1, 0)
	return next.AddDate(0, 0, -1)
}
