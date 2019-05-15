// Wrapper for `git describe` that adds --semver
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mroth/semverdesc"
	"github.com/mroth/semverdesc/describer"
)

var (
	// flags unique to us...
	// ...main
	path   = flag.String("path", ".", "path of git repo to describe, otherwise current workdir")
	legacy = flag.Bool("legacy", false, "display results similar to 'git describe --tags', e.g. not semver compliant")

	// flags compatible with git-describe...
	// ...search
	tags       = flag.Bool("tags", false, "use any tag, even unannotated")
	candidates = flag.Uint("candidates", 10, "consider `<n>` most recent tags")
	debug      = flag.Bool("debug", false, "debug search strategy on stderr")
	// ...formatting
	abbrev = flag.Uint("abbrev", 7, "use `<n>` digits to display SHA-1s")
	long   = flag.Bool("long", false, "always use long format")

	// Some potential additions to implement down the line if there is strong demand:
	// --match <pattern>     only consider tags matching <pattern>
	// --exclude <pattern>   do not consider tags matching <pattern>
	// --dirty[=<mark>]      append <mark> on dirty working tree (default: "-dirty") // see pflag.NoOptDefVal
)

func main() {
	flag.Parse()
	opts := describer.Options{
		Debug:      *debug,
		Tags:       *tags,
		Candidates: *candidates,
	}
	formatOpts := semverdesc.FormatOptions{
		Abbrev: *abbrev,
		Long:   *long,
	}

	commitish := flag.Arg(0)
	if commitish == "" {
		commitish = "HEAD"
	}
	d, err := describer.DescribeAtPath(*path, commitish, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v.\n", err)
		os.Exit(128) //??
	}

	if *legacy {
		fmt.Println(d.FormatLegacy(formatOpts))
	} else {
		fmt.Println(d.Format(formatOpts))
	}
}

// do a full comparison along with shelling out to git describe to compare its output
// func debugCompare() {
// 	localOpts := localgit.NewDescribeOptions().Set(func(opts *localgit.DescribeOptions) {
// 	})
// }
