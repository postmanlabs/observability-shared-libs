package akid

import "fmt"

// A service ID paired with its organization ID.
type QualifiedServiceID struct {
	OrganizationID OrganizationID `json:"organization_id"`
	ServiceID      ServiceID      `json:"service_id"`
}

func MakeQualifiedServiceID(organizationID OrganizationID, serviceID ServiceID) QualifiedServiceID {
	return QualifiedServiceID{
		OrganizationID: organizationID,
		ServiceID:      serviceID,
	}
}

// Qualifies the given learn session ID with this service ID.
func (serviceID QualifiedServiceID) QualifyLearnSessionID(learnSessionID LearnSessionID) QualifiedLearnSessionID {
	return MakeQualifiedLearnSessionID(serviceID.OrganizationID, serviceID.ServiceID, learnSessionID)
}

func (serviceID QualifiedServiceID) String() string {
	return fmt.Sprintf("%s/%s", serviceID.OrganizationID, serviceID.ServiceID)
}

// A learn session ID paired with its service ID and organization ID.
type QualifiedLearnSessionID struct {
	QualifiedServiceID
	LearnSessionID LearnSessionID `json:"learn_session_id"`
}

func MakeQualifiedLearnSessionID(organizationID OrganizationID, serviceID ServiceID, learnSessionID LearnSessionID) QualifiedLearnSessionID {
	return QualifiedLearnSessionID{
		QualifiedServiceID: MakeQualifiedServiceID(organizationID, serviceID),
		LearnSessionID:     learnSessionID,
	}
}

func (sessionID QualifiedLearnSessionID) String() string {
	return fmt.Sprintf("%s/%s", sessionID.QualifiedServiceID, sessionID.LearnSessionID)
}
