package spec_summary

import (
	"fmt"
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/spec_util"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
	vis "github.com/akitasoftware/akita-libs/visitors/http_rest"
)

func Summarize(spec *pb.APISpec) *Summary {
	v := specSummaryVisitor{
		methodSummary: &Summary{
			Authentications: make(map[string]int),
			HTTPMethods:     make(map[string]int),
			Paths:           make(map[string]int),
			Params:          make(map[string]int),
			Properties:      make(map[string]int),
			ResponseCodes:   make(map[string]int),
			DataFormats:     make(map[string]int),
			DataKinds:       make(map[string]int),
			DataTypes:       make(map[string]int),
		},
		summary: &Summary{
			Authentications: make(map[string]int),
			HTTPMethods:     make(map[string]int),
			Paths:           make(map[string]int),
			Params:          make(map[string]int),
			Properties:      make(map[string]int),
			ResponseCodes:   make(map[string]int),
			DataFormats:     make(map[string]int),
			DataKinds:       make(map[string]int),
			DataTypes:       make(map[string]int),
		},
	}
	vis.Apply(go_ast.POSTORDER, &v, spec)
	return v.summary
}

type specSummaryVisitor struct {
	vis.DefaultHttpRestSpecVisitor

	// Count occurrences within a single method.
	methodSummary *Summary

	// Count the number of methods in which each term occurs.
	summary *Summary
}

func (v *specSummaryVisitor) VisitMethod(_ vis.HttpRestSpecVisitorContext, m *pb.Method) bool {
	if meta := spec_util.HTTPMetaFromMethod(m); meta != nil {
		v.summary.HTTPMethods[strings.ToUpper(meta.GetMethod())] += 1
		v.summary.Paths[meta.GetPathTemplate()] += 1
	}

	// For each term that occurs at least once in this method, increment the
	// summary count by one.
	summaryPairs := []struct {
		dst  map[string]int
		src map[string]int
		kind string
	}{
		{dst: v.summary.Authentications, src: v.methodSummary.Authentications, kind: "authentications"},
		{dst: v.summary.HTTPMethods, src: v.methodSummary.HTTPMethods, kind: "http_methods"},
		{dst: v.summary.Paths, src: v.methodSummary.Paths, kind: "paths"},
		{dst: v.summary.Params, src: v.methodSummary.Params, kind: "params"},
		{dst: v.summary.Properties, src: v.methodSummary.Properties, kind: "properties"},
		{dst: v.summary.ResponseCodes, src: v.methodSummary.ResponseCodes, kind: "response_codes"},
		{dst: v.summary.DataFormats, src: v.methodSummary.DataFormats, kind: "data_formats"},
		{dst: v.summary.DataKinds, src: v.methodSummary.DataKinds, kind: "data_kinds"},
		{dst: v.summary.DataTypes, src: v.methodSummary.DataTypes, kind: "data_types"},
	}
	for _, summaryPair := range summaryPairs {
		for key, count := range summaryPair.src {
			if count > 0 {
				summaryPair.dst[key] += 1
			}
			delete(summaryPair.src, key)
		}
	}

	return true
}

func (v *specSummaryVisitor) VisitData(_ vis.HttpRestSpecVisitorContext, d *pb.Data) bool {
	// Handle auth vs params vs properties.
	if meta := spec_util.HTTPAuthFromData(d); meta != nil {
		v.methodSummary.Authentications[meta.Type.String()] += 1
	} else if meta := spec_util.HTTPPathFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else if meta := spec_util.HTTPQueryFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else if meta := spec_util.HTTPHeaderFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else if meta := spec_util.HTTPCookieFromData(d); meta != nil {
		v.methodSummary.Params[meta.Key] += 1
	} else {
		if s, ok := d.Value.(*pb.Data_Struct); ok {
			for k := range s.Struct.GetFields() {
				v.methodSummary.Properties[k] += 1
			}
		}
	}

	// Handle response codes.
	if meta := spec_util.HTTPMetaFromData(d); meta != nil {
		if meta.GetResponseCode() != 0 { // response code 0 means it's a request
			v.methodSummary.ResponseCodes[fmt.Sprintf("%d", meta.GetResponseCode())] += 1
		}
	}

	return true
}

func (v *specSummaryVisitor) VisitPrimitive(_ vis.HttpRestSpecVisitorContext, p *pb.Primitive) bool {
	for f := range p.GetFormats() {
		v.methodSummary.DataFormats[f] += 1
	}
	if k := p.GetFormatKind(); k != "" {
		v.methodSummary.DataKinds[k] += 1
	}
	v.methodSummary.DataTypes[spec_util.TypeOfPrimitive(p)] += 1
	return true
}
