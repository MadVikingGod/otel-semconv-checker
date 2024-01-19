// SPDX-License-Identifier: Apache-2.0

package servers

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
)

func newTestMatchDef(groups []string, ignore []string) matchDef {
	return matchDef{
		name:   regexp.MustCompile(`.*`),
		group:  groups,
		ignore: ignore,
	}
}

func Test_matchDef_isMatch(t *testing.T) {
	matchDefFilled := matchDef{
		name:  regexp.MustCompile(`foo`),
		attrs: map[string]string{"attr1": "val1"},
	}

	type args struct {
		name  string
		attrs []*v1.KeyValue
	}

	fooVal := args{
		name: "foo",
		attrs: []*v1.KeyValue{
			createKeyValue("attr1", "val1"),
		},
	}
	barVal := args{
		name: "bar",
		attrs: []*v1.KeyValue{
			createKeyValue("attr1", "val1"),
		},
	}
	fooNone := args{
		name: "foo",
		attrs: []*v1.KeyValue{
			createKeyValue("attr1", "none"),
		},
	}
	barNone := args{
		name: "bar",
		attrs: []*v1.KeyValue{
			createKeyValue("attr1", "none"),
		},
	}

	tests := []struct {
		name  string
		match matchDef
		args  args
		want  bool
	}{
		{
			name:  "name match, attrs match",
			match: matchDefFilled,
			args:  fooVal,
			want:  true,
		},
		{
			name:  "name doesn't match, attrs match",
			match: matchDefFilled,
			args:  barVal,
		},
		{
			name: "name nil, attrs match",
			match: matchDef{
				attrs: map[string]string{"attr1": "val1"},
			},
			args: fooVal,
			want: true,
		},
		{
			name:  "name matches and attrs doesn't match",
			match: matchDefFilled,
			args:  fooNone,
		},
		{
			name:  "name doesn't match, attrs doesn't match",
			match: matchDefFilled,
			args:  barNone,
		},
		{
			name: "name nil, attrs doesn't match",
			match: matchDef{
				attrs: map[string]string{"attr1": "val1"},
			},
			args: fooNone,
		},
		{
			name: "name mathes and attrs nil",
			match: matchDef{
				name: regexp.MustCompile(`foo`),
			},
			args: fooVal,
			want: true,
		},
		{
			name: "name doesn't match, attrs nil",
			match: matchDef{
				name: regexp.MustCompile(`foo`),
			},
			args: barVal,
		},
		{
			name: "Attrbutes match among multiple",
			match: matchDef{
				attrs: map[string]string{
					"attr1": "val1",
				},
			},
			args: args{
				name: "foo",
				attrs: []*v1.KeyValue{
					createKeyValue("stuff", "none"),
					createKeyValue("attr1", "val1"),
				},
			},
			want: true,
		},
		{
			name: "multi attributes needs all",
			match: matchDef{
				attrs: map[string]string{
					"attr1": "val1",
					"attr2": "val2",
				},
			},
			args: args{
				name: "foo",
				attrs: []*v1.KeyValue{
					createKeyValue("stuff", "none"),
					createKeyValue("attr1", "val1"),
				},
			},
		},
		{
			name: "multi attributes has all",
			match: matchDef{
				attrs: map[string]string{
					"attr1": "val1",
					"attr2": "val2",
				},
			},
			args: args{
				name: "foo",
				attrs: []*v1.KeyValue{
					createKeyValue("stuff", "none"),
					createKeyValue("attr2", "val2"),
					createKeyValue("attr1", "val1"),
				},
			},
			want: true,
		},
		{
			name: "empty value matches any",
			match: matchDef{
				attrs: map[string]string{
					"attr1": "",
				},
			},
			args: args{
				name: "foo",
				attrs: []*v1.KeyValue{
					createKeyValue("stuff", "none"),
				},
			},
		},
		{
			name: "empty value matches any",
			match: matchDef{
				attrs: map[string]string{
					"attr1": "",
				},
			},
			args: args{
				name: "foo",
				attrs: []*v1.KeyValue{
					createKeyValue("attr1", "val1"),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.match.isMatch(tt.args.name, tt.args.attrs))
		})
	}
}

func createKeyValue(key, value string) *v1.KeyValue {
	return &v1.KeyValue{
		Key:   key,
		Value: &v1.AnyValue{Value: &v1.AnyValue_StringValue{StringValue: value}},
	}
}
