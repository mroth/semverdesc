package localgit

import "fmt"

type DescribeOptions struct {
	// Describe the state of the working tree. When the working tree matches HEAD, the output
	// is the same as "git describe HEAD". If the working tree has local modification
	// "-dirty" is appended to it. If a repository is corrupt and Git cannot determine if
	// there is local modification, Git will error out, unless `--broken' is given, which
	// appends the suffix "-broken" instead.
	DirtyMark  string
	BrokenMark string

	// Instead of using only the annotated tags, use any ref found in refs/
	// namespace. This option enables matching any known branch, remote-tracking
	// branch, or lightweight tag.
	All bool

	// Instead of using only the annotated tags, use any tag found in
	// refs/tags namespace. This option enables matching a lightweight
	// (non-annotated) tag.
	Tags bool

	// Instead of finding the tag that predates the commit, find the tag that
	// comes after the commit, and thus contains it. Automatically implies
	// --tags.
	Contains bool

	// Instead of using the default 7 hexadecimal digits as the abbreviated
	// object name, use <n> digits, or as many digits as needed to form a
	// unique object name. An <n> of 0 will suppress long format, only
	// showing the closest tag.
	Abbrev uint

	// Instead of considering only the 10 most recent tags as candidates to
	// describe the input commit-ish consider up to <n> candidates.
	// Increasing <n> above 10 will take slightly longer but may produce a
	// more accurate result. An <n> of 0 will cause only exact matches to be
	// output.
	Candidates uint

	// Only output exact matches (a tag directly references the supplied
	// commit). This is a synonym for --candidates=0.
	ExactMatch bool

	// Verbosely display information about the searching strategy being employed to standard
	// error. The tag name will still be printed to standard out
	Debug bool

	// Always output the long format (the tag, the number of commits and the
	// abbreviated commit name) even when it matches a tag. This is useful
	// when you want to see parts of the commit object name in "describe"
	// output, even when the commit in question happens to be a tagged
	// version. Instead of just emitting the tag name, it will describe such
	// a commit as v1.2-0-gdeadbee (0th commit since tag v1.2 that points at
	// object deadbee....).
	Long bool

	// Only consider tags matching the given glob(7) pattern, excluding the
	// "refs/tags/" prefix. If used with --all, it also considers local
	// branches and remote-tracking references matching the pattern,
	// excluding respectively "refs/heads/" and "refs/remotes/" prefix;
	// references of other types are never considered. If given multiple
	// times, a list of patterns will be accumulated, and tags matching any
	// of the patterns will be considered. Use --no-match to clear and reset
	// the list of patterns.
	MatchPattern string

	// Do not consider tags matching the given glob(7) pattern, excluding the
	// "refs/tags/" prefix. If used with --all, it also does not consider
	// local branches and remote-tracking references matching the pattern,
	// excluding respectively "refs/heads/" and "refs/remotes/" prefix;
	// references of other types are never considered. If given multiple
	// times, a list of patterns will be accumulated and tags matching any of
	// the patterns will be excluded. When combined with --match a tag will
	// be considered when it matches at least one --match pattern and does
	// not match any of the --exclude patterns. Use --no-exclude to clear and
	// reset the list of patterns.
	ExcludePattern string

	// Show uniquely abbreviated commit object as fallback.
	Always bool

	// Follow only the first parent commit upon seeing a merge commit. This
	// is useful when you wish to not match tags on branches merged in the
	// history of the target commit.
	FirstParent bool
}

// Whereby default values for DescribeOptions differ from the zero-values,
// the defaults are listed here for ease of reading in godocs.
//
// To obtain a new DescribeOptions with all default values
const (
	DescribeOptionsDefaultAbbrev     = 7
	DescribeOptionsDefaultCandidates = 10
)

// NewDescribeOptions returns a DescribeOptions with all default values set.
func NewDescribeOptions() *DescribeOptions {
	return &DescribeOptions{
		Abbrev:     DescribeOptionsDefaultAbbrev,
		Candidates: DescribeOptionsDefaultCandidates,
	}
}

// Set takes a function that operates on a *DescribeOptions, executes it on this
// instance, and returns a pointer.
//
// This is a convenience function that allows you to instantiate
//
// Example:
//
//	opts := NewDescribeOptions().Set(func(o *DescribeOptions) { o.All = true })
//
// TODO: migrate this example to an actual godoc Example?
//
// In a perfect world, we would not need this and could instead either use
// default struct values or even better Option types (to handle the
// disambiguation between the zero value uints in DescribeOptions), but this is
// not a perfect world, it is Go! :shrug:
func (o *DescribeOptions) Set(f func(*DescribeOptions)) *DescribeOptions {
	f(o)
	return o
}

// func (o *DescribeOptions) Args() []string {
// 	return []string{"describe", o.Flags()} // + TARGET
// }

// Flags returns the minimal set of arg flags suitable to construct a git-describe
// command line invocation with those options.
//
// By minimal it is meant that only flags which set options that differ from the
// default behavior will be included.
func (o *DescribeOptions) Flags() []string {
	var args []string
	args = appendValueFlag(args, "--dirty", o.DirtyMark)
	args = appendValueFlag(args, "--broken", o.BrokenMark)
	args = appendToggleFlag(args, "--all", o.All)
	args = appendToggleFlag(args, "--tags", o.Tags)
	args = appendToggleFlag(args, "--contains", o.Contains)

	args = appendUintFlagNoDefaults(args,
		"--abbrev", o.Abbrev, DescribeOptionsDefaultAbbrev)
	args = appendUintFlagNoDefaults(args,
		"--candidates", o.Candidates, DescribeOptionsDefaultCandidates)

	args = appendToggleFlag(args, "--exact-match", o.ExactMatch)
	args = appendToggleFlag(args, "--debug", o.Debug)
	args = appendToggleFlag(args, "--long", o.Long)
	args = appendValueFlag(args, "--match", o.MatchPattern)
	args = appendValueFlag(args, "--exclude", o.ExcludePattern)
	args = appendToggleFlag(args, "--always", o.Always)
	args = appendToggleFlag(args, "--first-parent", o.FirstParent)
	return args
}

// append key=value flag IF not same as default, this appears to only apply to
// uint values for git-describe, so hard-coded that way for now instead of any
// Stringer.
func appendUintFlagNoDefaults(args []string, flag string, v, def uint) []string {
	if v != def {
		return append(args, fmt.Sprintf("%v=%d", flag, v))
	}
	return args
}

// append key=value flag IF non-empty
func appendValueFlag(args []string, flag string, v string) []string {
	if v != "" {
		return append(args, fmt.Sprintf("%v=%s", flag, v))
	}
	return args
}

// append toggle flag (boolean that defaults to false), IF non-zero
func appendToggleFlag(args []string, flag string, on bool) []string {
	if on {
		return append(args, flag)
	}
	return args
}
