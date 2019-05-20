package semverdesc

import "testing"

func TestDescribeResults_Format(t *testing.T) {
	tests := []struct {
		name string
		desc DescribeResults
		opts FormatOptions
		want string
	}{
		{
			name: "default",
			desc: DescribeResults{
				TagName: "v0.2.1",
				Ahead:   15,
				HashStr: "d71dd5072d51458a534ca7e0ec7c181d84754774",
			},
			opts: DefaultFormatOptions(),
			want: "v0.2.1+15.gd71dd50",
		},
		{
			name: "adjust abbrev",
			desc: DescribeResults{
				TagName: "v0.2.1",
				Ahead:   15,
				HashStr: "d71dd5072d51458a534ca7e0ec7c181d84754774",
			},
			opts: FormatOptions{
				Abbrev: 10,
			},
			want: "v0.2.1+15.gd71dd5072d",
		},
		// since we take hashes as a string instead of a full [20]byte, it is
		// highly possible for someone to give us a hex hash string which is
		// shorter than their requested Abbrev length.  If this is the case,
		// just use the shorter of the two, protecting against panic due to a
		// slice out of bounds.
		{
			name: "adjust abbrev overflow",
			desc: DescribeResults{
				TagName: "v0.2.1",
				Ahead:   15,
				HashStr: "d71dd5",
			},
			opts: FormatOptions{
				Abbrev: 20,
			},
			want: "v0.2.1+15.gd71dd5",
		},
		{
			name: "exact match",
			desc: DescribeResults{
				TagName: "v0.1.2",
				Ahead:   0,
				HashStr: "71dd5072d51458a534ca7e0ec7c181d84754774d",
			},
			opts: DefaultFormatOptions(),
			want: "v0.1.2",
		},
		{
			name: "exact match with long",
			desc: DescribeResults{
				TagName: "v0.1.2",
				Ahead:   0,
				HashStr: "71dd5072d51458a534ca7e0ec7c181d84754774d",
			},
			opts: FormatOptions{
				Long: true,
			},
			want: "v0.1.2+0.g71dd507",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.desc.Format(tc.opts); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
