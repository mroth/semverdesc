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
	dr, err := gitgo.Describe(repo, hash, &ggOpts)
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
