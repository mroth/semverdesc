package localgit

import (
	"reflect"
	"testing"
)

func TestDescribeOptions_Flags(t *testing.T) {
	tests := []struct {
		name string
		opts *DescribeOptions
		want []string
	}{
		{
			name: "single option",
			opts: NewDescribeOptions().Set(func(o *DescribeOptions) {
				o.All = true
			}),
			want: []string{"--all"},
		},
		{
			name: "multiple options",
			opts: NewDescribeOptions().Set(func(o *DescribeOptions) {
				o.All = true
				o.ExactMatch = true
			}),
			want: []string{"--all", "--exact-match"},
		},
		{
			name: "only options not matching defaults",
			opts: NewDescribeOptions().Set(func(o *DescribeOptions) {
				o.Abbrev = DescribeOptionsDefaultAbbrev
				o.Candidates = 3
			}),
			// could be {"--abbrev=7", "--candidates=3"}, but abbrev=7 is already default!
			want: []string{"--candidates=3"},
		},
		{
			name: "uint zero values but not default",
			opts: NewDescribeOptions().Set(func(o *DescribeOptions) {
				o.Abbrev = 0
				o.Candidates = 0
			}),
			want: []string{"--abbrev=0", "--candidates=0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.Flags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DescribeOptions.Flags() = %v, want %v", got, tt.want)
			}
		})
	}
}
