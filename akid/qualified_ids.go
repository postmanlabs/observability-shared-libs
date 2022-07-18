package akid

// A service ID paired with its organization ID.
type QualifiedServiceID struct {
	organizationID OrganizationID
	serviceID      ServiceID
}

// A learn session ID paired with its service ID and organization ID.
type QualifiedLearnSessionID struct {
	QualifiedServiceID
	learnSessionID LearnSessionID
}
