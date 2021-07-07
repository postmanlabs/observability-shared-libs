package time_span

import "time"

// A closed interval of time.
type TimeSpan struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func NewTimeSpan(start time.Time, end time.Time) *TimeSpan {
	if start.After(end) {
		start, end = end, start
	}

	return &TimeSpan{
		Start: start,
		End:   end,
	}
}

func (span TimeSpan) Duration() time.Duration {
	return span.End.Sub(span.Start)
}

// Determines whether the span includes the given query.
func (span TimeSpan) Includes(query time.Time) bool {
	return span.Start.Before(query) && query.Before(span.End) || span.Start.Equal(query) || span.End.Equal(query)
}
