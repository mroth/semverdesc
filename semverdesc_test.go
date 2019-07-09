package semverdesc

import "testing"

var testCases = []struct {
	name   string
	desc   DescribeResults
	opts   FormatOptions
	want   string
	legacy string
}{
	{
		name: "default",
		desc: DescribeResults{
			TagName:  "v0.2.1",
			Distance: 15,
			HashStr:  "d71dd5072d51458a534ca7e0ec7c181d84754774",
		},
		opts:   DefaultFormatOptions(),
		want:   "v0.2.1+15.gd71dd50",
		legacy: "v0.2.1-15-gd71dd50",
	},
	{
		name: "adjust abbrev",
		desc: DescribeResults{
			TagName:  "v0.2.1",
			Distance: 15,
			HashStr:  "d71dd5072d51458a534ca7e0ec7c181d84754774",
		},
		opts: FormatOptions{
			Abbrev: 10,
		},
		want:   "v0.2.1+15.gd71dd5072d",
		legacy: "v0.2.1-15-gd71dd5072d",
	},
	// since we take hashes as a string instead of a full [20]byte, it is
	// highly possible for someone to give us a hex hash string which is
	// shorter than their requested Abbrev length.  If this is the case,
	// just use the shorter of the two, protecting against panic due to a
	// slice out of bounds.
	{
		name: "adjust abbrev overflow",
		desc: DescribeResults{
			TagName:  "v0.2.1",
			Distance: 15,
			HashStr:  "d71dd5",
		},
		opts: FormatOptions{
			Abbrev: 20,
		},
		want:   "v0.2.1+15.gd71dd5",
		legacy: "v0.2.1-15-gd71dd5",
	},
	{
		name: "exact match",
		desc: DescribeResults{
			TagName:  "v0.1.2",
			Distance: 0,
			HashStr:  "71dd5072d51458a534ca7e0ec7c181d84754774d",
		},
		opts:   DefaultFormatOptions(),
		want:   "v0.1.2",
		legacy: "v0.1.2",
	},
	{
		name: "exact match with long",
		desc: DescribeResults{
			TagName:  "v0.1.2",
			Distance: 0,
			HashStr:  "71dd5072d51458a534ca7e0ec7c181d84754774d",
		},
		opts: FormatOptions{
			Long: true,
		},
		want:   "v0.1.2+0.g71dd507",
		legacy: "v0.1.2-0-g71dd507",
	},
	{
		name: "prerelease tag",
		desc: DescribeResults{
			TagName:  "v1.0.0-rc2",
			Distance: 0,
			HashStr:  "71dd5072d51458a534ca7e0ec7c181d84754774d",
		},
		opts:   DefaultFormatOptions(),
		want:   "v1.0.0-rc2",
		legacy: "v1.0.0-rc2",
	},
	{
		name: "prerelease tag with distance",
		desc: DescribeResults{
			TagName:  "v1.0.0-rc2",
			Distance: 2,
			HashStr:  "71dd5072d51458a534ca7e0ec7c181d84754774d",
		},
		opts:   DefaultFormatOptions(),
		want:   "v1.0.0-rc2+2.g71dd507",
		legacy: "v1.0.0-rc2-2-g71dd507",
	},
	{
		name: "exact dirty match without dirtymark (defaultopts)",
		desc: DescribeResults{
			TagName:  "v0.1.2",
			Distance: 0,
			HashStr:  "71dd5072d51458a534ca7e0ec7c181d84754774d",
			Dirty:    true,
		},
		opts:   DefaultFormatOptions(),
		want:   "v0.1.2",
		legacy: "v0.1.2",
	},
	{
		name: "exact dirty match with dirtymark",
		desc: DescribeResults{
			TagName:  "v0.1.2",
			Distance: 0,
			HashStr:  "71dd5072d51458a534ca7e0ec7c181d84754774d",
			Dirty:    true,
		},
		opts: FormatOptions{
			DirtyMark: "-dirty",
		},
		want:   "v0.1.2-dirty",
		legacy: "v0.1.2-dirty",
	},
	{
		name: "exact dirty match with dirtymark+long",
		desc: DescribeResults{
			TagName:  "v0.1.2",
			Distance: 0,
			HashStr:  "71dd5072d51458a534ca7e0ec7c181d84754774d",
			Dirty:    true,
		},
		opts: FormatOptions{
			Long:      true,
			DirtyMark: ".dirty",
		},
		want:   "v0.1.2+0.g71dd507.dirty",
		legacy: "v0.1.2-0-g71dd507.dirty",
	},
}

func TestDescribeResults_Format(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.desc.Format(tc.opts); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDescribeResults_FormatLegacy(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.desc.FormatLegacy(tc.opts); got != tc.legacy {
				t.Errorf("got %v, want %v", got, tc.legacy)
			}
		})
	}
}

func BenchmarkFormat(b *testing.B) {
	d := testCases[0].desc
	opts := DefaultFormatOptions()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Format(opts)
	}
}
