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
	buildVersion = "0.0.0-dev"
)

var (
	// flags compatible with git-describe...
	all         = pflag.Bool("all", false, "use any ref")
	tags        = pflag.Bool("tags", false, "use any tag, even unannotated")
	long        = pflag.Bool("long", false, "always use long format")
	firstParent = pflag.Bool("first-parent", false, "only follow first parent")
	abbrev      = pflag.Uint("abbrev", semverdesc.DefaultFormatAbbrev, "use `<n>` digits to display SHA-1s")
	exactMatch  = pflag.Bool("exact-match", false, "only output exact matches")
	candidates  = pflag.Uint("candidates", describer.DefaultCandidatesOption, "consider `<n>` most recent tags")
	match       = pflag.String("match", "", "only consider tags matching `<pattern>`")
	exclude     = pflag.String("exclude", "", "do not consider tags matching `<pattern>`")
	dirty       = pflag.String("dirty", "", "append `<mark>` on dirty working tree")

	// flags unique to us...
	path    = pflag.String("path", "", "path of git repo to describe (default $PWD)")
	legacy  = pflag.Bool("legacy", false, "format results like normal git describe")
	version = pflag.Bool("version", false, "display version information and exit")
)

func main() {
	pflag.ErrHelp = errors.New("")
	pflag.CommandLine.SortFlags = false
	pflag.CommandLine.Lookup("dirty").NoOptDefVal = "-dirty"
	pflag.CommandLine.MarkHidden("version")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: git semver-describe [<options>] [<commit-ish>]\n")
		fmt.Fprintf(os.Stderr, "   or: git semver-describe [<options>] --dirty\n\n")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if *version {
		fmt.Println("git-semver-describe version", buildVersion)
		os.Exit(0)
	}

	opts := describer.Options{
		Tags:           *tags,
		Candidates:     *candidates,
		MatchPattern:   *match,
		ExcludePattern: *exclude,
		All:            *all,
		ExactMatch:     *exactMatch,
		FirstParent:    *firstParent,
	}
	formatOpts := semverdesc.FormatOptions{
		Abbrev:    *abbrev,
		Long:      *long,
		DirtyMark: *dirty,
	}

	commitish := pflag.Arg(0)
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
