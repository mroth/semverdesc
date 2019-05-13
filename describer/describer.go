// Package describer is the glue that queries a git repository from the
// filesystem and turns the results back into something semverdesc expects.
//
// In current state, this is essentially a bridge between semverdesc and gitgo,
// so neither need to know about eachother. It's a little duplicative, but was
// extracted so people who just want semverdesc library functions dont need to
// bundle all the rest, which matters since gogit is so huge.
//
// This also enables us to keep our gitgo package closer to a pure extension of
// gogit and not build our specific needs in, eventually leading to something
// that might be able to be a PR on go-git project.
package describer

import (
	"github.com/mroth/semverdesc"
	"github.com/mroth/semverdesc/gitgo"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Options gitgo.DescribeOptions

func DescribePath(path string, commitish string, opts Options) (*semverdesc.DescribeResults, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Reference(plumbing.ReferenceName(commitish), true)
	if err != nil {
		return nil, err
	}
	return describe(repo, ref, opts)
}

func describe(repo *git.Repository, ref *plumbing.Reference, opts Options) (*semverdesc.DescribeResults, error) {
	ggOpts := gitgo.DescribeOptions(opts)
	dr, err := gitgo.Describe(repo, ref, &ggOpts)
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
		HashStr: dr.Reference.Hash().String(),
	}
}
