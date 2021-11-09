package http_rest

import (
	"fmt"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	. "github.com/akitasoftware/akita-libs/visitors"
	"github.com/akitasoftware/akita-libs/visitors/go_ast"
)

// This is the default implementation of VisitNodeChildren for DefaultSpecVisitorImpl.
// It is indirectly called by every Visit____Children that does not override the default method.
//
// We need to call the visitor manager on:
//   Every exported field in a struct, after updating the context with EnterStruct.
//   Every element of an array, after updating the context with EnterArray.
//   Every element of a map, after updating the context with EnterMap.
// Each call to visit may abort the operation.
//
// This implementation uses reflection as a fallback, but has specialized methods for common types.
// GRPC types use the fallback.
//
// It looks like the original implementation would call EnterNode/VisitChildren/LeaveNode on basic types.
// We won't be doing that, as specVisitor's enter/visitChildren/leave methods do not do anything with them.
//
func DefaultVisitIRChildren(ctx Context, vm VisitorManager, m interface{}) Cont {
	keepGoing := Continue

	// A check for m == nil doesn't work here, because m could be a typed pointer
	// to nil, which is different than interface{}(nil).

	switch node := m.(type) {
	case []*pb.Method:
		for i, m := range node {
			keepGoing = visitSliceMember(ctx, vm, m, i, node[i])
			if keepGoing != Continue {
				break
			}
		}
	case []*pb.Data:
		// Yes, this is identical to the one above
		for i, m := range node {
			keepGoing = visitSliceMember(ctx, vm, m, i, node[i])
			if keepGoing != Continue {
				break
			}
		}
	case map[string]*pb.Data:
		for k, v := range node {
			keepGoing = visitMapMember(ctx, vm, m, k, v)
			if keepGoing != Continue {
				break
			}
		}
	case map[string]*pb.ExampleValue:
		// Yes, this is identical to the one above
		for k, v := range node {
			keepGoing = visitMapMember(ctx, vm, m, k, v)
			if keepGoing != Continue {
				break
			}
		}
	case *pb.Witness:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Method", node.Method)
		}
	case *pb.MethodMeta_Http:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Http", node.Http)
		}
	case *pb.DataMeta_Grpc:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Grpc", node.Grpc)
		}
	case *pb.DataMeta_Http:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Http", node.Http)
		}
	case *pb.Data_Primitive:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Primitive", node.Primitive)
		}
	case *pb.Data_Struct:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Struct", node.Struct)
		}
	case *pb.Data_List:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "List", node.List)
		}
	case *pb.Data_Optional:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Optional", node.Optional)
		}
	case *pb.Data_Oneof:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Oneof", node.Oneof)
		}
	case *pb.AkitaAnnotations:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "FormatOption", node.FormatOption)
		}
	case *pb.FormatOption:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Format", node.Format)
		}
	case *pb.Primitive_BoolValue:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "BoolValue", node.BoolValue)
		}
	case *pb.Primitive_BytesValue:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "BytesValue", node.BytesValue)
		}
	case *pb.Primitive_StringValue:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "StringValue", node.StringValue)
		}
	case *pb.Primitive_Int32Value:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Int32Value", node.Int32Value)
		}
	case *pb.Primitive_Int64Value:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Int64Value", node.Int64Value)
		}
	case *pb.Primitive_Uint32Value:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Uint32Value", node.Uint32Value)
		}
	case *pb.Primitive_Uint64Value:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Uint64Value", node.Uint64Value)
		}
	case *pb.Primitive_DoubleValue:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "DoubleValue", node.DoubleValue)
		}
	case *pb.Primitive_FloatValue:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "FloatValue", node.FloatValue)
		}
	case *pb.MapData:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m,
				"Key", node.Key,
				"Value", node.Value,
			)
		}
	case *pb.Optional_None:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "None", node.None)
		}
	case *pb.Optional_Data:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Data", node.Data)
		}
	case *pb.HTTPMeta_Path:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Path", node.Path)
		}
	case *pb.HTTPMeta_Query:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Query", node.Query)
		}
	case *pb.HTTPMeta_Header:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Header", node.Header)
		}
	case *pb.HTTPMeta_Cookie:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Cookie", node.Cookie)
		}
	case *pb.HTTPMeta_Body:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Body", node.Body)
		}
	case *pb.HTTPMeta_Empty:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Empty", node.Empty)
		}
	case *pb.HTTPMeta_Auth:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Auth", node.Auth)
		}
	case *pb.HTTPMeta_Multipart:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Multipart", node.Multipart)
		}
	case *pb.Bool:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.Bytes:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.String:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.Int32:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.Int64:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.Uint32:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.Uint64:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.Float:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	case *pb.Double:
		if node != nil {
			keepGoing = visitStructMembers(ctx, vm, m, "Type", node.Type)
		}
	// These all have No non-basic types as children, so no further recursion is needed
	case *pb.MethodID: // FIXME: should we even enter this one? Not used by SpecVisitor
	case *pb.FormatOption_StringFormat: // only contains a string
	case *pb.None:
	case *pb.StringType:
	case *pb.BytesType:
	case *pb.Int32Type:
	case *pb.Int64Type:
	case *pb.Uint32Type:
	case *pb.Uint64Type:
	case *pb.FloatType:
	case *pb.DoubleType:
	case *pb.GRPCMeta:
	case *pb.GRPCMethodMeta:
	case *pb.ExampleValue:
	default:
		return go_ast.DefaultVisitChildren(ctx, vm, m)
	}
	return keepGoing
}

// Visit the struct members given by <field1> <value1> <field2> <value2>
func visitStructMembers(ctx Context, vm VisitorManager, inStruct interface{}, fields ...interface{}) Cont {
	keepGoing := Continue
	for i := 0; i < len(fields); i += 2 {
		field := fields[i].(string)
		value := fields[i+1]
		// Recurse into the member
		keepGoing := VisitInterface(ctx.EnterStruct(inStruct, field), vm, value)

		// Validate return value
		switch keepGoing {
		case Abort, Stop:
			return keepGoing
		case Continue:
			// fall through and get the next value, if any
		case SkipChildren:
			panic("VisitInterface returned SkipChildren")
		default:
			panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
		}
	}
	return keepGoing
}

func visitSliceMember(ctx Context, vm VisitorManager, inSlice interface{}, index int, value interface{}) Cont {
	// Recurse into the member
	keepGoing := VisitInterface(ctx.EnterArray(inSlice, index), vm, value)

	// Validate return value
	switch keepGoing {
	case Abort, Stop, Continue:
		return keepGoing
	case SkipChildren:
		panic("VisitInterface returned SkipChildren")
	default:
		panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
	}
}

func visitMapMember(ctx Context, vm VisitorManager, inSlice interface{}, key string, value interface{}) Cont {
	// Recurse into the member
	keepGoing := VisitInterface(ctx.EnterMapValue(inSlice, key), vm, value)

	// Validate return value
	switch keepGoing {
	case Abort, Stop, Continue:
		return keepGoing
	case SkipChildren:
		panic("VisitInterface returned SkipChildren")
	default:
		panic(fmt.Sprintf("Unknown Cont value: %d", keepGoing))
	}
}

// Visits the node m of any type, whose context is c. This should never return SkipChildren.
// Implements all the logic to enter, visit children, and leave, while obeying the visitor's returned
// Cont value.
// This is a copy of astVisitor's visit function.
func VisitInterface(c Context, vm VisitorManager, m interface{}) Cont {
	return go_ast.ApplyWithContext(vm, c, m)
}
