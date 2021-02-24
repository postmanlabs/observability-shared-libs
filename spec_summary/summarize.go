package spec_summary

import (
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/spec_util"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
	vis "github.com/akitasoftware/akita-libs/visitors/http_rest"
)

func Summarize(spec *pb.APISpec) *Summary {
	v := specSummaryVisitor{
		summary: &Summary{
			Authentications: make(map[string]struct{}),
			HTTPMethods:     make(map[string]struct{}),
			Paths:           make(map[string]struct{}),
			Params:          make(map[string]struct{}),
			Properties:      make(map[string]struct{}),
			ResponseCodes:   make(map[int32]struct{}),
			DataFormats:     make(map[string]struct{}),
			DataKinds:       make(map[string]struct{}),
			DataTypes:       make(map[string]struct{}),
		},
	}
	vis.Apply(go_ast.PREORDER, &v, spec)
	return v.summary
}

type specSummaryVisitor struct {
	vis.DefaultHttpRestSpecVisitor

	summary *Summary
}

func (v *specSummaryVisitor) VisitMethod(_ vis.HttpRestSpecVisitorContext, m *pb.Method) bool {
	if meta := spec_util.HTTPMetaFromMethod(m); meta != nil {
		v.summary.HTTPMethods[strings.ToUpper(meta.GetMethod())] = struct{}{}
		v.summary.Paths[meta.GetPathTemplate()] = struct{}{}
	}
	return true
}

func (v *specSummaryVisitor) VisitData(_ vis.HttpRestSpecVisitorContext, d *pb.Data) bool {
	// Handle auth vs params vs properties.
	if meta := spec_util.HTTPAuthFromData(d); meta != nil {
		v.summary.Authentications[meta.Type.String()] = struct{}{}
	} else if meta := spec_util.HTTPPathFromData(d); meta != nil {
		v.summary.Params[meta.Key] = struct{}{}
	} else if meta := spec_util.HTTPQueryFromData(d); meta != nil {
		v.summary.Params[meta.Key] = struct{}{}
	} else if meta := spec_util.HTTPHeaderFromData(d); meta != nil {
		v.summary.Params[meta.Key] = struct{}{}
	} else if meta := spec_util.HTTPCookieFromData(d); meta != nil {
		v.summary.Params[meta.Key] = struct{}{}
	} else {
		if s, ok := d.Value.(*pb.Data_Struct); ok {
			for k := range s.Struct.GetFields() {
				v.summary.Properties[k] = struct{}{}
			}
		}
	}

	// Handle response codes.
	if meta := spec_util.HTTPMetaFromData(d); meta != nil {
		if meta.GetResponseCode() != 0 { // response code 0 means it's a request
			v.summary.ResponseCodes[meta.GetResponseCode()] = struct{}{}
		}
	}

	return true
}

func (v *specSummaryVisitor) VisitPrimitive(_ vis.HttpRestSpecVisitorContext, p *pb.Primitive) bool {
	for f := range p.GetFormats() {
		v.summary.DataFormats[f] = struct{}{}
	}
	if k := p.GetFormatKind(); k != "" {
		v.summary.DataKinds[k] = struct{}{}
	}
	v.summary.DataTypes[spec_util.TypeOfPrimitive(p)] = struct{}{}
	return true
}
