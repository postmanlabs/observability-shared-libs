package time_span

import "time"

// A closed interval of time.
type TimeSpan struct {
	start time.Time `json:"start"`
	end   time.Time `json:"end"`
}

func NewTimeSpan(start time.Time, end time.Time) *TimeSpan {
	if start.After(end) {
		start, end = end, start
	}

	return &TimeSpan{
		start: start,
		end:   end,
	}
}

func (span TimeSpan) Start() time.Time {
	return span.start
}

func (span TimeSpan) End() time.Time {
	return span.end
}

func (span TimeSpan) Duration() time.Duration {
	return span.end.Sub(span.start)
}

// Determines whether the span includes the given query.
func (span TimeSpan) Includes(query time.Time) bool {
	return span.start.Before(query) && query.Before(span.end) || span.start.Equal(query) || span.end.Equal(query)
}
