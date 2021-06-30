package tags

type CreatedBy = string

// Valid values for the XAkitaCreatedBy tag.
const (
	// Designates a spec that was automatically created by a schedule.
	CreatedBySchedule CreatedBy = "schedule"

	// Designates a spec that was automatically created as an aggregate ("big")
	// model.
	CreatedByBigModel CreatedBy = "aggregator"
)
