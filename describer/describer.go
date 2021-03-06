// Package describer is the glue that queries a git repository from the
// filesystem and turns the results back into something semverdesc expects.
package describer

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/mroth/semverdesc"
	"github.com/mroth/semverdesc/localgit"
)

// Options that can adjust the  describe operation.
type Options struct {
	// Instead of using only the annotated tags, use any ref found in refs/
	// namespace. This option enables matching any known branch, remote-tracking
	// branch, or lightweight tag.
	All bool

	// Instead of using only the annotated tags, use any tag found in
	// refs/tags namespace. This option enables matching a lightweight
	// (non-annotated) tag.
	Tags bool

	// Instead of considering only the 10 most recent tags as candidates to
	// describe the input commit-ish consider up to <n> candidates.
	// Increasing <n> above 10 will take slightly longer but may produce a
	// more accurate result. An <n> of 0 will cause only exact matches to be
	// output.
	Candidates uint

	// Only output exact matches (a tag directly references the supplied
	// commit). This is a synonym for --candidates=0.
	ExactMatch bool

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

	// Follow only the first parent commit upon seeing a merge commit. This
	// is useful when you wish to not match tags on branches merged in the
	// history of the target commit.
	FirstParent bool
}

/*
There are a few git describe related flags not currently implemented in
describer.

TODO: maybe --debug?

WONTFIX: --always, unless strongly requested. Having a fallback like this is
incompatible with --long and thus makes parsing more complicated (we could also
just implement ourselves.)

WONTFIX: --broken, unless requested.  I dont think I've ever seen this used and
complicates parsing.

WONTFIX: --contains. This is a totally different weird format, e.g. v0.3.0~10 is
ten commits prior to v0.3.0. I don't believe it currently fits in with this use
case so lets remove it instead of trying to figure out how to parse and
translate.
*/

// DefaultCandidatesOption is the suggested default value for *Options.Candidates
const DefaultCandidatesOption = uint(10)

// Describe attempts to perform a git describe operation on a git repository
// located at path, with an optional commit-ish describing the target to
// describe. For the default case (describing HEAD), set the commitish as the
// zero value.
//
// Note that the returned error may be of type exec.ExitError if there was an
// error condition returned from the underlying git describe command. You can
// check for this to handle the output differently!
func Describe(path, commitish string, opts Options) (*semverdesc.DescribeResults, error) {
	cmd := buildCmd(path, commitish, opts)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parsePDescribe(output)
}

// Options used to get predictable formatting out of the underlying localgit
// describe operation, so that we can parse it in a reasoned way.
const (
	pAbbrev    = uint(40)
	pDirtyMark = "-dirty"
	pLong      = true
)

// buildCmd creates the localgit shell command to do the describe and return
// our predictable output.
func buildCmd(path, commitish string, opts Options) *exec.Cmd {
	gdOpts := localgit.DescribeOptions{
		// DescribeOptions for the search are passed along directly
		All:            opts.All,
		Tags:           opts.Tags,
		Candidates:     opts.Candidates,
		ExactMatch:     opts.ExactMatch,
		MatchPattern:   opts.MatchPattern,
		ExcludePattern: opts.ExcludePattern,
		FirstParent:    opts.FirstParent,
		// On the other hand, formatting options we set explicitly to make the
		// output predictable and parse it later.
		Abbrev:    pAbbrev,
		Long:      pLong,
		DirtyMark: pDirtyMark,
	}

	args := []string{"describe"}
	args = append(args, gdOpts.Flags()...)
	if commitish != "" {
		args = append(args, commitish)
	}
	cmd := exec.Command("git", args...)
	// git describe assumes working directory, so we just set that based on path :-)
	cmd.Dir = path
	return cmd
}

// regex to match git describe output when predictable format options applied
var pdescRegex = regexp.MustCompile(
	fmt.Sprintf(`^(.+)-(\d+)-g([0-9a-f]{%d})(%s)?$`, pAbbrev, pDirtyMark),
)

// parsePDescribe parses our "predictable" describe as defined by our
// expected describe output options.
func parsePDescribe(output []byte) (*semverdesc.DescribeResults, error) {
	output = bytes.TrimSuffix(output, []byte("\n"))
	if len(output) == 0 {
		return nil, errors.New("received empty output")
	}
	match := pdescRegex.FindSubmatch(output)
	if match == nil {
		return nil, errors.New("unable to match: [" + string(output) + "]")
	}

	// if we ended in `-dirty`, last match group will not be empty
	dirty := len(match[4]) != 0
	// sha is the pAbbrev hex chars prior to that, but after the `-g`
	sha := match[3]
	// the distance is a series of digits
	digits := string(match[2])
	distance, err := strconv.Atoi(digits)
	if err != nil {
		return nil, errors.New("could not parse distance: " + digits)
	}
	// everything before was the tag
	tag := match[1]

	return &semverdesc.DescribeResults{
		TagName:  string(tag),
		Distance: uint(distance),
		HashStr:  string(sha),
		Dirty:    dirty,
	}, nil
}
