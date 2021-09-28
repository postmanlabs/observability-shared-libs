package time_span

import "time"

// Return the earlier of the two times
func MinTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	} else {
		return b
	}
}

// Return the later of the two times
func MaxTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return b
	} else {
		return a
	}
}

type BaseInterval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// A closed interval of time.
// Empty intervals are represented by having Start > End.
type ClosedInterval BaseInterval

func NewClosedInterval(start time.Time, end time.Time) ClosedInterval {
	if start.After(end) {
		start, end = end, start
	}

	return ClosedInterval{
		Start: start,
		End:   end,
	}
}

func (span ClosedInterval) Empty() bool {
	return span.End.Before(span.Start)
}

func (span ClosedInterval) Duration() time.Duration {
	if span.Empty() {
		return time.Duration(0)
	} else {
		return span.End.Sub(span.Start)
	}
}

// Determines whether the span includes the given query.
func (span ClosedInterval) Includes(query time.Time) bool {
	return !(span.Empty() || query.Before(span.Start) || query.After(span.End))
}

// Extend an interval by "delta" in each direction (thus it gets 2*delta longer.)
// If it was empty, expand around the start point.
func (t ClosedInterval) Expand(delta time.Duration) ClosedInterval {
	if t.Empty() {
		return ClosedInterval{
			Start: t.Start.Add(time.Duration(-1) * delta),
			End:   t.Start.Add(delta),
		}
	} else {
		return ClosedInterval{
			Start: t.Start.Add(time.Duration(-1) * delta),
			End:   t.End.Add(delta),
		}
	}
}

// Return the portion of this interval that lies entirely within another closed interval.
// Intersection with an open interval or half-open interval is not provided because the result might
// either be open or closed, and that's a pain to deal with (the point of splitting them out was
// to stop confusing the type of interval involved and be able to statically type-check it!)
func (t ClosedInterval) Intersect(t2 ClosedInterval) ClosedInterval {
	return ClosedInterval{
		Start: MaxTime(t.Start, t2.Start), // later of the two starts
		End:   MinTime(t.End, t2.End),     // earlier of the two ends
	}
}

// Determine whether there is an overlap between two closed intervals
func (t1 ClosedInterval) Overlaps(t2 ClosedInterval) bool {
	// There's some time C such that start1 <= C <= end1 and start2 <= C <= end2
	// iff start1 <= end2 AND start2 <= end1.
	return (t1.Start.Before(t2.End) || t1.Start.Equal(t2.End)) &&
		(t2.Start.Before(t1.End) || t2.Start.Equal(t1.End)) &&
		(!t1.Empty() && !t2.Empty())
}

// Return the smallest interval containing the start and end points of the
// two intervals, even if they are empty.
// That is, the maximum of end times and the minimum of start times.
func (t ClosedInterval) Combine(t2 ClosedInterval) ClosedInterval {
	return ClosedInterval{
		Start: MinTime(t.Start, t2.Start), // earlier of the two starts
		End:   MaxTime(t.End, t2.End),     // later of the two ends
	}
}

// An half-open interval [start,end)
// Empty intervals are represented by having Start >= End.
type HalfOpenInterval BaseInterval

func NewHalfOpenInterval(start time.Time, end time.Time) HalfOpenInterval {
	return HalfOpenInterval{
		Start: start,
		End:   end,
	}
}

func (span HalfOpenInterval) Empty() bool {
	return span.End.Before(span.Start) || span.End.Equal(span.Start)
}

func (span HalfOpenInterval) Duration() time.Duration {
	if span.Empty() {
		return time.Duration(0)
	} else {
		return span.End.Sub(span.Start)
	}
}

// Determines whether the interval includes the given query.
func (span HalfOpenInterval) Includes(query time.Time) bool {
	return (span.Start.Equal(query) || span.Start.Before(query)) && query.Before(span.End)
}

func (t1 HalfOpenInterval) Overlaps(t2 HalfOpenInterval) bool {
	// There's some time C such that start1 <= C < end1 and start2 <= C < end2
	// iff start1 < end2 AND start2 < end1, and both intervals are nonempty.
	return t1.Start.Before(t2.End) && t2.Start.Before(t1.End) && !t1.Empty() && !t2.Empty()
}

// Returns true if t2 is entirely contained within t1 (i.e., a subset)
func (t1 HalfOpenInterval) Contains(t2 HalfOpenInterval) bool {
	if t1.Empty() {
		return false
	}
	if t2.Empty() {
		return true
	}
	return (t1.Start.Before(t2.Start) || t1.Start.Equal(t2.Start)) &&
		(t1.End.After(t2.End) || t1.End.Equal(t2.End))
}

// Extend an interval by "delta" in each direction (thus it gets 2*delta longer.)
// If empty, return an interval of size 2 * delta around the start point.
func (t HalfOpenInterval) Expand(delta time.Duration) HalfOpenInterval {
	if t.Empty() {
		return HalfOpenInterval{
			Start: t.Start.Add(time.Duration(-1) * delta),
			End:   t.Start.Add(delta),
		}
	} else {
		return HalfOpenInterval{
			Start: t.Start.Add(time.Duration(-1) * delta),
			End:   t.End.Add(delta),
		}
	}
}

// Return the intersection of two half-open intervals.
// (i.e., trim this interval to lie entirely within another.)
// May result in an interval with end <= start, which should be treated as empty.
func (t HalfOpenInterval) Intersect(t2 HalfOpenInterval) HalfOpenInterval {
	return HalfOpenInterval{
		Start: MaxTime(t.Start, t2.Start), // later of the two starts
		End:   MinTime(t.End, t2.End),     // earlier of the two ends
	}
}

// Return the smallest interval containing the start and end points of the
// two intervals, even if they are empty.
// That is, the maximum of end times and the minimum of start times,
//
// The important use case for handling events is that the zero-length interval
// [a,a) can be combined with a non-empty interval [b,c) not including a
// to produce [a,c) or [b,a).
func (t HalfOpenInterval) Combine(t2 HalfOpenInterval) HalfOpenInterval {
	return HalfOpenInterval{
		Start: MinTime(t.Start, t2.Start), // earlier of the two starts
		End:   MaxTime(t.End, t2.End),     // later of the two ends
	}
}

type TimeSpan = HalfOpenInterval

func NewTimeSpan(start time.Time, end time.Time) *TimeSpan {
	if start.After(end) {
		start, end = end, start
	}

	return &TimeSpan{
		Start: start,
		End:   end,
	}
}
