// Package semverdesc defines a standard format for... TODO
package semverdesc

import "fmt"

type DescribeResults struct {
	// The name of the matched tag
	TagName string
	// Number of commits ahead
	Ahead uint
	// The SHA hash of the commit-ish used for the describe, converted to a
	// string. For compatibility, we do not validate(?).
	HashStr string
}

type FormatOptions struct {
	// Instead of using the default 7 hexadecimal digits as the abbreviated
	// object name, use <n> digits, or as many digits as needed to form a unique
	// object name. An <n> of 0 will suppress long format, only showing the
	// closest tag.
	//
	// Re: "or as many digits as needed to form a unique object name." this is not
	// currently implemented probably impossible.
	Abbrev uint
	Long   bool
	// DirtyMark
}

func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		Abbrev: 7,
		Long:   false,
	}
}

// String blah blah DEscribrResults with default options
func (dr *DescribeResults) String() string {
	return dr.Format(DefaultFormatOptions())
}

func (dr *DescribeResults) Format(opts FormatOptions) string {
	if dr.Ahead == 0 || opts.Abbrev == 0 {
		return dr.TagName
	}
	return semverLongFormat(dr, opts)
}

// approximation of...
func (dr *DescribeResults) FormatLegacy(opts FormatOptions) string {
	if dr.Ahead == 0 || opts.Abbrev == 0 {
		return dr.TagName
	}
	return legacyLongFormat(dr, opts)
}

func semverLongFormat(dr *DescribeResults, opts FormatOptions) string {
	abbrev := effectiveAbbrev(dr, opts)
	return fmt.Sprintf("%v+%v.g%v", dr.TagName, dr.Ahead, dr.HashStr[:abbrev])
}

func legacyLongFormat(dr *DescribeResults, opts FormatOptions) string {
	abbrev := effectiveAbbrev(dr, opts)
	return fmt.Sprintf("%v-%v-g%v", dr.TagName, dr.Ahead, dr.HashStr[:abbrev])
}

// we never want to slice outside of index, and since we are allowing arbitrary
// strings for the hashStr value here,
//
// This isn't actually an issue for our internal usage where we get the hash
// from a [20]byte using go-git plumbing, but havig a looser string based
// definition in our API allows for easier re-use by others.
func effectiveAbbrev(dr *DescribeResults, opts FormatOptions) uint {
	if hashStrLen := uint(len(dr.HashStr)); hashStrLen < opts.Abbrev {
		return hashStrLen
	}
	return opts.Abbrev
}

// TODO: allow parsing old strings? urgh... could be messy
// func FromLegacy(in string) (*DescribeResults, error) {}
