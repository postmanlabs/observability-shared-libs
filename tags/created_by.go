package tags

type CreatedBy = string

// Valid values for the XAkitaCreatedBy tag.
const (
	// Designates a spec that was automatically created by a schedule.
	CreatedBySchedule CreatedBy = "schedule"

	// Designates a spec that was automatically created as a "big model".
	CreatedByBigModel CreatedBy = "big model"
)
