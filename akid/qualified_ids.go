package akid

// A service ID paired with its organization ID.
type QualifiedServiceID struct {
	OrganizationID OrganizationID
	ServiceID      ServiceID
}

// A learn session ID paired with its service ID and organization ID.
type QualifiedLearnSessionID struct {
	QualifiedServiceID
	LearnSessionID LearnSessionID
}
