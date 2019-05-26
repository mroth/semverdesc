// Package semverdesc defines a standardized method for formatting `git
// describe` results that is compliant with Semantic Versioning 2.0.
// (https://semver.org)
package semverdesc

import "fmt"

// DescribeResults are the structured results  from a `git describe` operation
// on a commit, ready to be formatted.
type DescribeResults struct {
	// The name of the matched tag
	TagName string
	// Number of commits ahead
	Ahead uint
	// The SHA hash of the commit-ish used for the describe, converted to a
	// string. For compatibility, we do not validate(?).
	HashStr string
	// Dirty is true if the working tree has local modifications.
	Dirty bool
}

// FormatOptions control the output when formatting a DescribeResults.
//
// Note that `0` is a valid value for certain options, and has semantic
// meaning within git describe that differs greatly from what would be expected
// in a zero value. Thus, if you are not explicitly defining all struct values,
// you will likely want to start with DefaultFormatOptions().
type FormatOptions struct {
	// Instead of using the default 7 hexadecimal digits as the abbreviated
	// object name, use <n> digits. An <n> of 0 will suppress long format, only
	// showing the closest tag.
	//
	// Whereas official git describe has a "...or as many digits as needed to
	// form a unique object name." caveat, this is not currently implemented
	// here.
	Abbrev uint
	// Always use long format, even if exact match.
	Long bool
	// Describe the state of the working tree. When the working tree matches
	// HEAD, the output is the same as "git describe HEAD". If the working tree
	// has local modification DirtyMark is appended to it.
	DirtyMark string
}

// Defaults which differ from their zero values
const (
	DefaultFormatAbbrev = uint(7)
)

// DefaultFormatOptions returns the default FormatOptions.
func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		Abbrev: DefaultFormatAbbrev,
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
		return dr.TagName + dirtySuffix(dr, opts)
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
		return dr.TagName + dirtySuffix(dr, opts)
	}
	return legacyLongFormat(dr, opts)
}

func semverLongFormat(dr *DescribeResults, opts FormatOptions) string {
	abbrev := effectiveAbbrev(dr, opts)
	return fmt.Sprintf("%v+%v.g%v%v",
		dr.TagName, dr.Ahead, dr.HashStr[:abbrev], dirtySuffix(dr, opts))
}

func legacyLongFormat(dr *DescribeResults, opts FormatOptions) string {
	abbrev := effectiveAbbrev(dr, opts)
	return fmt.Sprintf("%v-%v-g%v%v",
		dr.TagName, dr.Ahead, dr.HashStr[:abbrev], dirtySuffix(dr, opts))
}

// dirtySuffix returns the DirtyMark suffix if *DescribeResults are both Dirty
// and FormatOptions has a nonzero DirtyMark.
func dirtySuffix(dr *DescribeResults, opts FormatOptions) string {
	if dr.Dirty && opts.DirtyMark != "" {
		return opts.DirtyMark
	}
	return ""
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
