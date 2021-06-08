package spec_util

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type MeldedMethod interface {
	GetMethod() *pb.Method

	GetArgs() map[string]*pb.Data
	SetArgs(map[string]*pb.Data)

	GetResponses() map[string]*pb.Data
	SetResponses(map[string]*pb.Data)

	Has4xxOnly() bool
	SetHas4xxOnly(bool)
	Clone() MeldedMethod

	// Melds src into this instance, resolving conflicts using oneof. Assumes that this and src are for the same endpoint.
	//
	// Responses are always melded. But because they are likely to contain bogus data, requests that result in 4xx response codes are ignored where possible:
	//   - If both src and this contain only 4xx responses, then requests are melded.
	//   - Otherwise, if src contains only 4xx responses, then its requests are ignored.
	//   - Otherwise, if this contains only 4xx responses, then its requests are replaced with requests from src.
	//   - Otherwise, neither src nor this contain only 4xx responses, and requests are melded.
	Meld(src MeldedMethod) error
}

type meldedMethod struct {
	method *pb.Method

	// Indicates whether the method has only requests that received a 4xx response code.
	has4xxOnly bool
}

var _ MeldedMethod = (*meldedMethod)(nil)

func NewMeldedMethod(method *pb.Method) MeldedMethod {
	return &meldedMethod{
		method:     method,
		has4xxOnly: hasOnly4xxResponses(method),
	}
}

func (m *meldedMethod) GetMethod() *pb.Method {
	return m.method
}

func (m *meldedMethod) GetArgs() map[string]*pb.Data {
	return m.method.GetArgs()
}

func (m *meldedMethod) SetArgs(args map[string]*pb.Data) {
	m.method.Args = args
}

func (m *meldedMethod) GetResponses() map[string]*pb.Data {
	return m.method.GetResponses()
}

func (m *meldedMethod) SetResponses(responses map[string]*pb.Data) {
	m.method.Responses = responses
}

func (m *meldedMethod) Has4xxOnly() bool {
	return m.has4xxOnly
}

func (m *meldedMethod) SetHas4xxOnly(b bool) {
	m.has4xxOnly = b
}

func (m *meldedMethod) Clone() MeldedMethod {
	return &meldedMethod{
		method:     proto.Clone(m.method).(*pb.Method),
		has4xxOnly: m.has4xxOnly,
	}
}

func (dst *meldedMethod) Meld(src MeldedMethod) error {
	if !src.Has4xxOnly() && dst.has4xxOnly {
		// Replace dst requests with src requests.
		dst.method.Args = src.GetArgs()
		dst.has4xxOnly = false
	} else if src.Has4xxOnly() && !dst.has4xxOnly {
		// Ignore src requests.
	} else {
		// Meld requests.
		if dst.method.Args == nil {
			dst.method.Args = src.GetArgs()
		} else if err := meldTopLevelDataMap(dst.method.Args, src.GetArgs()); err != nil {
			return errors.Wrap(err, "failed to meld arg map")
		}
		dst.has4xxOnly = dst.has4xxOnly && src.Has4xxOnly()
	}

	// Meld responses.
	if dst.method.Responses == nil {
		dst.method.Responses = src.GetResponses()
	} else if err := meldTopLevelDataMap(dst.method.Responses, src.GetResponses()); err != nil {
		return errors.Wrap(err, "failed to meld response map")
	}

	return nil
}

// Determines whether a given method has only 4xx response codes. Returns true if the method has at least one response and all response codes are 4xx.
func hasOnly4xxResponses(method *pb.Method) bool {
	responses := method.GetResponses()
	if len(responses) == 0 {
		return false
	}

	for _, response := range responses {
		responseCode := response.Meta.GetHttp().GetResponseCode()
		if responseCode < 400 || responseCode >= 500 {
			return false
		}
	}

	return true
}
