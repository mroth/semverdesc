// Package describer is the glue that queries a git repository from the
// filesystem and turns the results back into something semverdesc expects.
package describer

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/mroth/semverdesc"
	"github.com/mroth/semverdesc/localgit"
)

type Options struct {
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
}

// Note that the returned error may be of type exec.ExitError if there was an error
// condition returned from the underlying git describe command. You can check for
// this to handle the output differently!
func Describe(path, commitish string, opts Options) (*semverdesc.DescribeResults, error) {
	gdOpts := localgit.DescribeOptions{
		// DescribeOptions for the search are passed along directly
		All:            opts.All,
		Tags:           opts.Tags,
		Contains:       opts.Contains,
		Candidates:     opts.Candidates,
		ExactMatch:     opts.ExactMatch,
		MatchPattern:   opts.MatchPattern,
		ExcludePattern: opts.ExcludePattern,
		// On the other hand, formatting options we set explicitly to make the
		// output predictable and parse it later.
		Abbrev:    40,
		Long:      true,
		DirtyMark: "-dirty",
	}

	args := []string{"describe"}
	args = append(args, gdOpts.Flags()...)
	if commitish != "" {
		args = append(args, commitish)
	}
	cmd := exec.Command("git", args...)
	// git describe assumes working directory, so we just set that based on path :-)
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parsePDescribe(output)
}

var pdescRegex = regexp.MustCompile(`^(.+)-(\d+)-g([0-9a-f]{40})(-dirty)?$`)

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
	_ = len(match[4]) != 0 // TODO: assign dirty
	// sha is the 40 hex chars prior to that, but after the `-g`
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
		TagName: string(tag),
		Ahead:   uint(distance),
		HashStr: string(sha),
	}, nil
}

/*
// In current state, this is essentially a bridge between semverdesc and gitgo,
// so neither need to know about eachother. It's a little duplicative, but was
// extracted so people who just want semverdesc library functions dont need to
// bundle all the rest, which matters since gogit is so huge.
//
// This also enables us to keep our gitgo package closer to a pure extension of
// gogit and not build our specific needs in, eventually leading to something
// that might be able to be a PR on go-git project.

type Options gitgo.DescribeOptions

func DescribeAtPath(path string, commitish string, opts Options) (*semverdesc.DescribeResults, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(commitish))
	if err != nil {
		return nil, err
	}
	return describe(repo, hash, opts)
}

func describe(repo *git.Repository, hash *plumbing.Hash, opts Options) (*semverdesc.DescribeResults, error) {
	ggOpts := gitgo.DescribeOptions(opts)
	dr, err := gitgo.DescribeCommit(repo, hash, &ggOpts)
	if err != nil {
		return nil, err
	}
	res := convert(dr)
	return &res, nil
}

func convert(dr *gitgo.DescribeResults) semverdesc.DescribeResults {
	return semverdesc.DescribeResults{
		TagName: dr.Tag.Name().Short(),
		Ahead:   uint(dr.Distance),
		HashStr: dr.Hash.String(),
	}
}
*/
