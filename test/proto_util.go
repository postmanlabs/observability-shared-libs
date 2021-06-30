package test

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func loadProtoFromFileOrDie(path string, m proto.Message) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to read proto from %s: %v", path, err))
	}
	if err := proto.UnmarshalText(string(content), m); err != nil {
		panic(fmt.Sprintf("failed to unmarshal proto from %s: %v", path, err))
	}
}

func LoadMethodFromFileOrDie(path string) *pb.Method {
	m := &pb.Method{}
	loadProtoFromFileOrDie(path, m)
	return m
}

func LoadDataFromFileOrDie(path string) *pb.Data {
	d := &pb.Data{}
	loadProtoFromFileOrDie(path, d)
	return d
}

func LoadDataTemplateFromFileOrDie(path string) *pb.DataTemplate {
	t := &pb.DataTemplate{}
	loadProtoFromFileOrDie(path, t)
	return t
}

func LoadMethodTemplateFromFileOrDie(path string) *pb.MethodTemplate {
	t := &pb.MethodTemplate{}
	loadProtoFromFileOrDie(path, t)
	return t
}

func LoadSequenceFromFileOrDie(path string) *pb.Sequence {
	s := &pb.Sequence{}
	loadProtoFromFileOrDie(path, s)
	return s
}

func LoadMethodDataRefFromFileOrDie(path string) *pb.MethodDataRef {
	r := &pb.MethodDataRef{}
	loadProtoFromFileOrDie(path, r)
	return r
}

func LoadDataRefFromFileOrDie(path string) *pb.DataRef {
	r := &pb.DataRef{}
	loadProtoFromFileOrDie(path, r)
	return r
}

func LoadAPISpecFromFileOrDie(path string) *pb.APISpec {
	a := &pb.APISpec{}
	loadProtoFromFileOrDie(path, a)
	return a
}

func LoadMethodCallsFromFileOrDie(path string) *pb.MethodCalls {
	a := &pb.MethodCalls{}
	loadProtoFromFileOrDie(path, a)
	return a
}

func LoadMethodsFromFileOrDie(path string) []*pb.Method {
	spec := LoadAPISpecFromFileOrDie(path)
	return spec.Methods
}

func LoadWitnessFromFileOrDie(path string) *pb.Witness {
	w := &pb.Witness{}
	loadProtoFromFileOrDie(path, w)
	return w
}

var LoadWitnessFromFileOrDile = LoadWitnessFromFileOrDie
