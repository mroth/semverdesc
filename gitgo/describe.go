// Package gitgo extends go-git to have rudimentary `git describe`
// functionality, which is not currently built-in to the library.
//
// Work in progress in the core library (appears abandoned?!):
// https://github.com/src-d/go-git/pull/816
//
// See also the Rust bindings for libgit2 which I believe perhaps exposes a
// better interface for this:
// https://docs.rs/git2
package gitgo

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// DescribeResults are Describe as defined in PR#816 but with formatting related
// results removed, as output formatting can (and should?) be handled separately
// from the Describe results themselves.
type DescribeResults struct {
	// Reference being described
	Reference *plumbing.Reference
	// Tag of the describe object
	Tag *plumbing.Reference
	// Distance to the tag object in commits
	Distance int
	// Dirty string to append
	// Dirty string
	// Dirty bool
}

// BUNCH OF LEFTOVER FORMATTING BOILERPLATE FROM VARIOUS EXPERIMENTS HERE.
// CURRENTLY IM NOT ACTUALLY EVEN DOING FORMATTING HERE (since semverdesc
// handles that, see over there for some real formatting code), BUT MAY WANT TO
// IMPLEMENT IN FUTURE TO GET THIS CLOSER TO SOMETHING THAT COULD BECOME A CORE
// LIB PR?
/*
func (dr *DescribeResults) String() string {
	if dr.Distance == 0 {
		return dr.Tag.Name().Short()
	}
	return dr.longString()
}

// LongString is equivalent of the output model of `--long` when passed to `git
// describe`.
//
// Always output the long format (the tag, the number of commits and the
// abbreviated commit name) even when it matches a tag. This is useful when you
// want to see parts of the commit object name in "describe" output, even when
// the commit in question happens to be a tagged version. Instead of just
// emitting the tag name, it will describe such a commit as v1.2-0-gdeadbee (0th
// commit since tag v1.2 that points at object deadbee....).
//
// The number of additional commits is the number of commits which would be
// displayed by "git log v1.0.4..parent".
//
// The hash suffix is "-g" + 7-char abbreviation for the tip commit of parent.
// The "g" prefix stands for "git" and is used to allow describing the version
// of a software depending on the SCM the software is managed with. This is
// useful in an environment where people may use different SCMs.
func (dr *DescribeResults) longString() string {
	return fmt.Sprintf("%v-%v-g%v",
		dr.Tag.Name().Short(), dr.Distance, dr.Reference.Hash().String()[:7])
}

// DescribeFormatOptions can be used to customize how a description is formatted.
type DescribeFormatOptions struct {
	// The value is the lower bound for the length of the abbreviated string, and the default is 7.
	Abbrev uint
	// Sets whether or not the long format is used even when a shorter name could be used.
	Long bool
	// If the workdir is dirty and this is set, this string will be appended to the description string.
	DirtySuffix string
}

func (o *DescribeFormatOptions) Validate() error {
	if o.Abbrev == 0 {
		o.Abbrev = 7
	}
	return nil
}
*/

type DescribeOptions struct {
	Debug      bool
	Tags       bool
	Candidates uint
	// All  bool
	// Contains bool
	// MatchPattern string
	// ExcludePattern string
}

func (o *DescribeOptions) Validate() error {
	if o.Candidates == 0 {
		o.Candidates = 10
	}
	return nil
}

// Describe is the `git describe` as WIP in the potentially abandoned PR#816.
//
// Few minor modifications:
//  - Catch some additional error conditions instead of panicking.
//  - Deal with some things as uint instead of int to avoid invalid negative values
func Describe(r *git.Repository, ref *plumbing.Reference, opts *DescribeOptions) (*DescribeResults, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Describes through the commit log ordered by commit time seems to be the best approximation to
	// git describe.
	commitIterator, err := r.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}

	// To query tags we create a temporary map.
	tagIterator, err := r.Tags()
	if err != nil {
		return nil, err
	}
	tags := make(map[plumbing.Hash]*plumbing.Reference)
	tagIterator.ForEach(func(t *plumbing.Reference) error {
		if to, err := r.TagObject(t.Hash()); err == nil {
			tags[to.Target] = t
		} else {
			tags[t.Hash()] = t
		}
		return nil
	})
	tagIterator.Close()
	if len(tags) < 1 {
		return nil, errors.New("no names found, cannot describe anything")
	}

	// The search looks for a number of suitable candidates in the log (specified through the options)
	type describeCandidate struct {
		ref       *plumbing.Reference
		annotated bool
		distance  int
	}
	var candidates []*describeCandidate
	var candidatesFound uint
	var count = -1
	var lastCommit *object.Commit

	if opts.Debug {
		fmt.Fprintf(os.Stderr, "searching to describe %v\n", ref.Name())
	}

	for {
		var candidate = &describeCandidate{annotated: false}

		err = commitIterator.ForEach(func(commit *object.Commit) error {
			lastCommit = commit
			count++
			if tagReference, ok := tags[commit.Hash]; ok {
				delete(tags, commit.Hash)
				candidate.ref = tagReference
				hash := tagReference.Hash()
				if !bytes.Equal(commit.Hash[:], hash[:]) {
					candidate.annotated = true
				}
				return storer.ErrStop
			}
			return nil
		})

		if candidate.annotated || opts.Tags {
			if candidatesFound < opts.Candidates {
				candidate.distance = count
				candidates = append(candidates, candidate)
			}
			candidatesFound++
		}

		if candidatesFound > opts.Candidates || len(tags) == 0 {
			break
		}
	}

	if opts.Debug {
		for _, c := range candidates {
			var description = "lightweight"
			if c.annotated {
				description = "annotated"
			}
			fmt.Fprintf(os.Stderr, " %-11s %8d %v\n", description, c.distance, c.ref.Name().Short())
		}
		fmt.Fprintf(os.Stderr, "traversed %v commits\n", count)
		if candidatesFound > opts.Candidates {
			fmt.Fprintf(os.Stderr, "more than %v tags found; listed %v most recent\n",
				opts.Candidates, len(candidates))
		}
		fmt.Fprintf(os.Stderr, "gave up search at %v\n", lastCommit.Hash.String())
	}

	// TODO(mroth): rationalize this error with above candidate search.
	// Currently this code appears to find N candidates, but only uses the first
	// one no matter what. So now we just check if at least one candidate exists
	// to bail gracefully instead of panicking on mis-slice.
	//
	// However, why is the candidate results built up to begin with? Assuming it has
	// to deal with unimplemented search features, so keeping for now in case I want
	// to add those features, need to check C git source code and see what it does.
	if len(candidates) < 1 {
		return nil, fmt.Errorf("no tags can describe %v", ref.Hash())
	}
	// Error git describe sometimes returns in this case:
	//
	// 		fatal: No annotated tags can describe 'd71dd5072d51458a534ca7e0ec7c181d84754774'.
	// 		However, there were unannotated tags: try --tags.
	//
	// To support that, looks like this core logic would have to be modified to track unaccepted
	// candidates as well.

	return &DescribeResults{
		ref,
		candidates[0].ref,
		candidates[0].distance,
	}, nil
}

// My original implementation below:
//
// tagMap is a map of commit hashs to known tag refs
// type tagMap map[plumbing.Hash]*plumbing.Reference

// func tagMapForRepo(r *git.Repository) (tagMap, error) {
// 	tags := make(tagMap)
// 	tagIterator, err := r.Tags()
// 	if err != nil {
// 		return nil, err
// 	}
// 	tagIterator.ForEach(func(t *plumbing.Reference) error {
// 		tags[t.Hash()] = t
// 		return nil
// 	})
// 	return tags, nil
// }
//
// func MyDescribe(r *git.Repository, ref *plumbing.Reference) (*DescribeResults, error) {
// 	tags, err := tagMapForRepo(r)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(tags) < 1 {
// 		return nil, errors.New("no names found, cannot describe anything")
// 	}

// 	commits, err := r.Log(&git.LogOptions{
// 		From:  ref.Hash(),
// 		Order: git.LogOrderCommitterTime,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	var count uint
// 	var tagRef *plumbing.Reference
// 	commits.ForEach(func(c *object.Commit) error {
// 		if t, ok := tags[c.Hash]; ok {
// 			tagRef = t
// 			return storer.ErrStop
// 		}
// 		count++
// 		return nil
// 	})
// 	// TODO: handle case if doesnt reach tag.. what does git describe do?

// 	return &DescribeResults{
// 		Tag:   tagRef.Name(),
// 		Ahead: count,
// 		Hash:  ref.Hash(),
// 	}, nil
// }
