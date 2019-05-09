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
)

// TODO: describes HEAD, does not yet take a COMMITISH
func DescribePath(path string) (*semverdesc.DescribeResults, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return describeRepo(repo)
}

func describeRepo(repo *git.Repository) (*semverdesc.DescribeResults, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	opts := gitgo.DescribeOptions{
		Debug:      false,
		Tags:       true,
		Candidates: 0,
	}
	dr, err := gitgo.Describe(repo, head, &opts)
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
