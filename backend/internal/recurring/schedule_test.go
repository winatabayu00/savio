package recurring

import (
	"testing"
	"time"
)

func d(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func ds(v []time.Time) []string {
	out := make([]string, len(v))
	for i := range v {
		out[i] = v[i].Format("2006-01-02")
	}
	return out
}

func expectDates(t *testing.T, freq string, start time.Time, end *time.Time, horizon time.Time, want []string) {
	t.Helper()
	got := ds(nextDates(freq, start, end, horizon))
	if len(got) != len(want) {
		t.Fatalf("freq %s got %v want %v", freq, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("freq %s got %v want %v", freq, got, want)
		}
	}
}

func TestScheduleDaily(t *testing.T) {
	expectDates(t, "DAILY", d("2026-01-01"), nil, d("2026-01-05"),
		[]string{"2026-01-01", "2026-01-02", "2026-01-03", "2026-01-04", "2026-01-05"})
}

func TestScheduleWeekly(t *testing.T) {
	expectDates(t, "WEEKLY", d("2026-01-05"), nil, d("2026-01-19"),
		[]string{"2026-01-05", "2026-01-12", "2026-01-19"})
}

func TestScheduleMonthlyClampsShortMonth(t *testing.T) {
	expectDates(t, "MONTHLY", d("2026-01-31"), nil, d("2026-03-31"),
		[]string{"2026-01-31", "2026-02-28", "2026-03-31"})
}

func TestScheduleMonthlyLeapYear(t *testing.T) {
	expectDates(t, "MONTHLY", d("2024-01-31"), nil, d("2024-03-31"),
		[]string{"2024-01-31", "2024-02-29", "2024-03-31"})
}

func TestScheduleMonthEnd(t *testing.T) {
	expectDates(t, "MONTH_END", d("2026-01-10"), nil, d("2026-03-31"),
		[]string{"2026-01-31", "2026-02-28", "2026-03-31"})
}

func TestScheduleHonoursEndDate(t *testing.T) {
	end := d("2026-01-12")
	expectDates(t, "DAILY", d("2026-01-10"), &end, d("2026-01-31"),
		[]string{"2026-01-10", "2026-01-11", "2026-01-12"})
}