// Wrapper for `git describe` that adds semverdesc output as the default.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/mroth/semverdesc"
	"github.com/mroth/semverdesc/describer"
	"github.com/spf13/pflag"
)

var (
	// flags unique to us...
	// ...main
	path   = pflag.String("path", "", "path of git repo to describe (default $PWD)")
	legacy = pflag.Bool("legacy", false, "format results like normal git describe")

	// flags compatible with git-describe...
	// ...search
	tags       = pflag.Bool("tags", false, "use any tag, even unannotated")
	candidates = pflag.Uint("candidates", 10, "consider `<n>` most recent tags")
	match      = pflag.String("match", "", "only consider tags matching `<pattern>`")
	exclude    = pflag.String("exclude", "", "do not consider tags matching `<pattern>`")
	// debug      = pflag.Bool("debug", false, "debug search strategy on stderr")
	// ...formatting
	abbrev = pflag.Uint("abbrev", 7, "use `<n>` digits to display SHA-1s")
	long   = pflag.Bool("long", false, "always use long format")
	dirty  = pflag.String("dirty", "", "append `<mark>` on dirty working tree")
)

func main() {
	pflag.ErrHelp = errors.New("")
	// https://github.com/spf13/pflag#setting-no-option-default-values-for-flags
	pflag.Lookup("dirty").NoOptDefVal = "-dirty"
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: git semver-describe [<options>] [<commit-ish>]\n")
		fmt.Fprintf(os.Stderr, "   or: git semver-describe [<options>] --dirty\n\n")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	opts := describer.Options{
		Tags:           *tags,
		Candidates:     *candidates,
		MatchPattern:   *match,
		ExcludePattern: *exclude,
		// All:            false,
		// Contains:       false,
		// ExactMatch:     false,
	}
	formatOpts := semverdesc.FormatOptions{
		Abbrev:    *abbrev,
		Long:      *long,
		DirtyMark: *dirty,
	}

	commitish := pflag.Arg(0)
	// if commitish == "" {
	// 	commitish = "HEAD"
	// }
	d, err := describer.Describe(*path, commitish, opts)
	if err != nil {
		// if was underlying git describe error, pass it along exactly
		if exiterr, ok := err.(*exec.ExitError); ok {
			fmt.Fprint(os.Stderr, string(exiterr.Stderr))
			os.Exit(exiterr.ExitCode())
		}
		// otherwise, handle as an error
		log.Fatal(err)
		os.Exit(1)
	}

	if *legacy {
		fmt.Println(d.FormatLegacy(formatOpts))
	} else {
		fmt.Println(d.Format(formatOpts))
	}
}
