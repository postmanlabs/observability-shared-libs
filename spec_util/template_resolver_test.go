package spec_util

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/akitasoftware/akita-libs/test"
)

func TestExtractValueFromTemplate(t *testing.T) {
	testCases := []struct {
		name             string
		dataTemplateFile string
		dataRefFile      string
		expectedDataFile string
		expectErr        bool
	}{
		{
			name:             "constant",
			dataTemplateFile: "testdata/template_resolver/value_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/primitive_ref.pb.txt",
			expectedDataFile: "testdata/template_resolver/string_value_1.pb.txt",
		},
		{
			name:             "const struct template - full ref",
			dataTemplateFile: "testdata/template_resolver/const_struct_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/full_struct_ref.pb.txt",
			expectedDataFile: "testdata/template_resolver/const_struct.pb.txt",
		},
		{
			name:             "const struct template - field ref",
			dataTemplateFile: "testdata/template_resolver/const_struct_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/struct_field_ref.pb.txt",
			expectedDataFile: "testdata/template_resolver/int64_value_1.pb.txt",
		},
		{
			name:             "const struct template - non-existent struct field ref",
			dataTemplateFile: "testdata/template_resolver/const_struct_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/non_existent_struct_field_ref.pb.txt",
			expectErr:        true,
		},
		{
			name:             "non-const struct template - full ref error",
			dataTemplateFile: "testdata/template_resolver/non_const_struct_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/struct_field_ref.pb.txt",
			expectErr:        true,
		},
		{
			name:             "const list template - full ref",
			dataTemplateFile: "testdata/template_resolver/const_list_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/full_list_ref.pb.txt",
			expectedDataFile: "testdata/template_resolver/const_list.pb.txt",
		},
		{
			name:             "const list template - elem ref",
			dataTemplateFile: "testdata/template_resolver/const_list_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/list_elem_ref.pb.txt",
			expectedDataFile: "testdata/template_resolver/int64_value_1.pb.txt",
		},
		{
			name:             "const list template - bad index elem ref",
			dataTemplateFile: "testdata/template_resolver/const_list_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/bad_index_list_elem_ref.pb.txt",
			expectErr:        true,
		},
		{
			name:             "non-const list template - full ref error",
			dataTemplateFile: "testdata/template_resolver/non_const_list_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/full_list_ref.pb.txt",
			expectErr:        true,
		},
		{
			name:             "recursive ref not allowed",
			dataTemplateFile: "testdata/template_resolver/ref_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/full_list_ref.pb.txt", // doesn't matter
			expectErr:        true,
		},
		{
			name:             "optional value",
			dataTemplateFile: "testdata/template_resolver/optional_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/primitive_ref.pb.txt",
			expectedDataFile: "testdata/template_resolver/string_value_1.pb.txt",
		},
		{
			name:             "optional none treated as error",
			dataTemplateFile: "testdata/template_resolver/optional_none_template.pb.txt",
			dataRefFile:      "testdata/template_resolver/full_list_ref.pb.txt", // doesn't matter
			expectErr:        true,
		},
	}

	for _, c := range testCases {
		dt := test.LoadDataTemplateFromFileOrDie(c.dataTemplateFile)
		ref := test.LoadDataRefFromFileOrDie(c.dataRefFile)
		v, err := ExtractValueFromTemplate(dt, ref)
		if err != nil {
			if !c.expectErr {
				t.Errorf("[%s] unexpected error: %v", c.name, err)
			}
		} else {
			if c.expectErr {
				t.Errorf("[%s] expected error, didn't get one", c.name)
			} else {
				expectedValue := test.LoadDataFromFileOrDie(c.expectedDataFile)
				if diff := cmp.Diff(expectedValue, v, protocmp.Transform()); diff != "" {
					t.Errorf("[%s] found diff in value: %s", c.name, diff)
				}
			}
		}
	}
}
