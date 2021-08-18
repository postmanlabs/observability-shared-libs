package spec_util

import (
	"errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func HTTPPathFromTemplate(dt *pb.DataTemplate) *pb.HTTPPath {
	if m, ok := dt.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Path); ok {
			return l.Path
		}
	}
	return nil
}

func HTTPQueryFromTemplate(dt *pb.DataTemplate) *pb.HTTPQuery {
	if m, ok := dt.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Query); ok {
			return l.Query
		}
	}
	return nil
}

func HTTPHeaderFromTemplate(dt *pb.DataTemplate) *pb.HTTPHeader {
	if m, ok := dt.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Header); ok {
			return l.Header
		}
	}
	return nil
}

func HTTPCookieFromTemplate(dt *pb.DataTemplate) *pb.HTTPCookie {
	if m, ok := dt.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Cookie); ok {
			return l.Cookie
		}
	}
	return nil
}

func HTTPBodyFromTemplate(dt *pb.DataTemplate) *pb.HTTPBody {
	if m, ok := dt.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Body); ok {
			return l.Body
		}
	}
	return nil
}

func HTTPPathFromData(d *pb.Data) *pb.HTTPPath {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Path); ok {
			return l.Path
		}
	}
	return nil
}

func HTTPQueryFromData(d *pb.Data) *pb.HTTPQuery {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Query); ok {
			return l.Query
		}
	}
	return nil
}

func HTTPHeaderFromData(d *pb.Data) *pb.HTTPHeader {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Header); ok {
			return l.Header
		}
	}
	return nil
}

func HTTPCookieFromData(d *pb.Data) *pb.HTTPCookie {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Cookie); ok {
			return l.Cookie
		}
	}
	return nil
}

func HTTPBodyFromData(d *pb.Data) *pb.HTTPBody {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Body); ok {
			return l.Body
		}
	}
	return nil
}

func HTTPEmptyFromData(d *pb.Data) *pb.HTTPEmpty {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Empty); ok {
			return l.Empty
		}
	}
	return nil
}

func HTTPAuthFromData(d *pb.Data) *pb.HTTPAuth {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Auth); ok {
			return l.Auth
		}
	}
	return nil
}

func HTTPMultipartFromData(d *pb.Data) *pb.HTTPMultipart {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		if l, ok := m.Http.Location.(*pb.HTTPMeta_Multipart); ok {
			return l.Multipart
		}
	}
	return nil
}

func HTTPMetaFromData(d *pb.Data) *pb.HTTPMeta {
	if m, ok := d.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
		return m.Http
	}
	return nil
}

func HTTPMetaFromMethod(m *pb.Method) *pb.HTTPMethodMeta {
	if m, ok := m.GetMeta().GetMeta().(*pb.MethodMeta_Http); ok {
		return m.Http
	}
	return nil
}

// Extract responses returned under a successful status code
func HTTPSuccessResponses(m *pb.Method) map[string]*pb.Data {
	switch m.GetId().GetApiType() {
	case pb.ApiType_HTTP_REST:
		results := make(map[string]*pb.Data)
		for k, data := range m.Responses {
			if m, ok := data.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
				statusCode := m.Http.GetResponseCode()
				if 200 <= statusCode && statusCode <= 299 {
					// This excludes -1, which is the "default" response in OpenAPI that
					// is used to group together error responses.
					// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#responses-object
					results[k] = data
				}
			}
		}
		return results
	default:
		return m.Responses
	}
}

// Extract the response code from the first response (meaningful for witnesses,
// which contain only at most one response.)
func HTTPResponseCode(m *pb.Method) (int32, error) {
	for _, data := range m.Responses {
		if m, ok := data.GetMeta().GetMeta().(*pb.DataMeta_Http); ok {
			return m.Http.GetResponseCode(), nil
		}
	}
	return 0, errors.New("No HTTP response code found")
}
