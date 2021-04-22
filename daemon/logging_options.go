package daemon

import "github.com/akitasoftware/akita-libs/akid"

type LoggingOptions struct {
	// The trace to which logged events should be associated.
	TraceName string              `json:"trace_name"`
	TraceID   akid.LearnSessionID `json:"trace_id"`

	// The service ID to which the trace belongs.
	ServiceID akid.ServiceID `json:"service_id"`

	// A number in the range [0,1], indicating the fraction of events to log.
	SamplingRate float32 `json:"sampling_rate"`

	// Whether third-party trackers should be filtered from the trace before
	// being sent to the cloud.
	FilterThirdPartyTrackers bool `json:"filter_third_party_trackers"`
}

func NewLoggingOptions(traceName string, traceID akid.LearnSessionID, serviceID akid.ServiceID, samplingRate float32, filterThirdPartyTrackers bool) *LoggingOptions {
	return &LoggingOptions{
		TraceName:                traceName,
		ServiceID:                serviceID,
		TraceID:                  traceID,
		SamplingRate:             samplingRate,
		FilterThirdPartyTrackers: filterThirdPartyTrackers,
	}
}

type ActiveTraceDiff struct {
	ActivatedTraces   []LoggingOptions      `json:"activated_traces"`
	DeactivatedTraces []akid.LearnSessionID `json:"deactivated_traces"`
}

func NewActiveTraceDiff(activatedTraces []LoggingOptions, deactivatedTraces []akid.LearnSessionID) *ActiveTraceDiff {
	return &ActiveTraceDiff{
		ActivatedTraces:   activatedTraces,
		DeactivatedTraces: deactivatedTraces,
	}
}

func (diff ActiveTraceDiff) Size() int {
	return len(diff.ActivatedTraces) + len(diff.DeactivatedTraces)
}
