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

// DefaultFormatOptions returns the default FormatOptions.
func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		Abbrev: 7,
		Long:   false,
	}
}

// String implements Stringer and is the equivalent of Format with
// DefaultFormatOptions.
func (dr *DescribeResults) String() string {
	return dr.Format(DefaultFormatOptions())
}

// Format returns the semver describe string utilizing given FormatOptions.
//
// Note that the zero value FormatOptions are not defaults in all cases, so if
// you to modify default formatting begin with DefaultFormatOptions().
func (dr *DescribeResults) Format(opts FormatOptions) string {
	if (dr.Ahead == 0 || opts.Abbrev == 0) && !opts.Long {
		return dr.TagName
	}
	return semverLongFormat(dr, opts)
}

// FormatLegacy returns the describe string in the same format that good old
// fashioned `git describe` would. This is provided for comparative reasons.
//
// Note that the zero value FormatOptions are not defaults in all cases, so if
// you to modify default formatting begin with DefaultFormatOptions().
func (dr *DescribeResults) FormatLegacy(opts FormatOptions) string {
	if (dr.Ahead == 0 || opts.Abbrev == 0) && !opts.Long {
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

// Calculate the "effective" Abbrev which may differ from the one that is passed
// as an option.
//
// 1) We never want to slice outside of index, and since we are allowing
// arbitrary strings for the hashStr value here, a requested Abbrev that is
// longer than what is possible could cause problems.
//
// This isn't actually an issue for our internal usage where we get the hash
// from a [20]byte using go-git plumbing, but having a looser string based
// definition in our library API allows for easier re-use by others.
//
// Because of this convenience, it is highly possible for someone to give us a
// hex hash string which is shorter than their requested Abbrev length.  If this
// is the case, just use the shorter of the two, protecting against panic due to
// a slice out of bounds.
//
// 2) Abbrev=0 (effectively what --short would be if git describe had resonable
// UX) and Long are incompatible options. If we get them, let Long win since
// Abbrev=0 was probably a lazy zero value.
func effectiveAbbrev(dr *DescribeResults, opts FormatOptions) uint {
	if hashStrLen := uint(len(dr.HashStr)); hashStrLen < opts.Abbrev {
		return hashStrLen
	}
	if opts.Abbrev == 0 && opts.Long {
		return DefaultFormatOptions().Abbrev
	}
	return opts.Abbrev
}

// TODO: allow parsing old strings? urgh... could be messy
// func FromLegacy(in string) (*DescribeResults, error) {}
