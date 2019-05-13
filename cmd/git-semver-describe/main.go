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
	abbrev = flag.Uint("abbrev", 7, "use `<n>` digits to display SHA-1s")
	path   = flag.String("path", ".", "path of git repo to describe, otherwise current workdir")
	legacy = flag.Bool("legacy", false, "display results similar to 'git describe --tags', e.g. not semver compliant")
	long   = flag.Bool("long", false, "always use long format")
	// tags   = flag.Bool("tags", true, "use any tag, even unannotated")

	// Some potential additions to implement down the line if there is strong demand:
	// --match <pattern>     only consider tags matching <pattern>
	// --exclude <pattern>   do not consider tags matching <pattern>
	// --dirty[=<mark>]      append <mark> on dirty working tree (default: "-dirty")
)

func main() {
	flag.Parse()
	// TODO: os.args is commitish, not path
	d, err := describer.DescribePath(*path, "HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v.\n", err)
		os.Exit(128) //??
	}
	formatOpts := semverdesc.FormatOptions{
		Abbrev: *abbrev,
		Long:   *long,
	}
	fmt.Println(d.FormatLegacy(formatOpts))
	fmt.Println(d.Format(formatOpts))
}

// do a full comparison along with shelling out to git describe to compare its output
// func debugCompare() {
// 	localOpts := localgit.NewDescribeOptions().Set(func(opts *localgit.DescribeOptions) {
// 	})
// }
