package tags

type ScheduledModelType = string

// Valid values for the XAkitaScheduledModelType tag.
const (
	// Designates a spec that was created as a "large model".  We show the
	// latest large model on the API Model page in the Akita app.
	ScheduledLargeModel ScheduledModelType = "large-model"

	// Designates a spec that was created as a large diffing model.  The
	// most recent large diffing model is used when computing model
	// differences.
	ScheduledLargeDiffingModel ScheduledModelType = "large-diffing-model"
	
	// Designates a spec that was created as a small diffing model.  The
	// most recent small diffing model is used when computing model
	// differences.
	ScheduledSmallDiffingModel ScheduledModelType = "small-diffing-model"
)
