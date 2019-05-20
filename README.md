# git-semver-describe :dart:

Extends `git describe` to return [Semantic Versioning v2.0](https://semver.org)
compatible strings:


```shell
$ git describe --tags
v0.2.1-15-gd71dd50

$ git semver-describe --tags
v0.2.1+15.gd71dd50

$ git semver-describe --tags --legacy
v0.2.1-15-gd71dd50
```

## Usage

For the most part, this is a drop-in replacement for `git describe`, with nearly
identical option syntax. The official git-describe has [tons of
options](https://git-scm.com/docs/git-describe), some of which are not supported
here, but the goal is to eventually cover the entire subset of options that are
potentially useful when using git tags for semantic versioning.

```shell
$ git semver-describe --help
usage: git semver-describe [<options>] [<commit-ish>]

      --abbrev <n>       use <n> digits to display SHA-1s (default 7)
      --candidates <n>   consider <n> most recent tags (default 10)
      --debug            debug search strategy on stderr
      --legacy           format results like normal git describe
      --long             always use long format
      --path string      path of git repo to describe (default $PWD)
      --tags             use any tag, even unannotated
```

## Installation

TODO

## API

There is also a Go library encapsulating a lot of this functionality, for more
information see the GoDocs.

## Detailed Discussion

_:warning: Warning: this is likely only interesting to you if really care about
semantic versioning..._

Firstly familiarize yourself with actual semver specification, this discussion
likely won't make much sense otherwise.

**git semver-describe** essentially takes the following philosophy: only tagged
refs are part of semver version precedence (including for pre-release versions),
everything else goes in the build metadata field.

How do these strings differ?

Here are some of the current issues in integrating `git describe` with semantic
versioning git tags:

- Git describe output would increment the pre-release patch version. But
  "prelease" versions come before a specific release. During early development,
  you don't necessarily know if the next release will be a major/minor/patch
  release.
- By definition, ***
- The output of git describe looks close to valid semver to the naked eye, but
  does not actually adhere to the specification.
- Once you tag an actual pre-release version (e.g. ), git describe's output
  drifts even further off from the specification. See the example below.

Consider the following sequence in both git describe and semver-describe:

| `git semver-describe`   | `git describe`           | comments                        |
|-------------------------|--------------------------|---------------------------------|
|                `v0.8.3` |                 `v0.8.3` | TAG: currently released version |
|     `v0.8.3+1.g1a2b3c4` |     `v0.8.3-1-g1a2b3c4`¹ | one commit later                |
|     `v0.8.3+2.g2b3c4d5` |      `v0.8.3-2-g2b3c4d5` | two commits later               |
|            `v0.9.0-rc1` |             `v0.9.0-rc1` | TAG: pre-minor-release          |
| `v0.9.0-rc1+1.gd3adb33` | `v0.9.0-rc1-1-gd3adb33`² | one commit later                |
|                `v0.9.0` |                 `v0.9.0` | TAG: minor-release made         |
<small>

1. This "describe" version is now something that comes _before_ the previous
   release in semver ordering.
2. This doesn't parse in SemVer v2.0 at all, but if it did, a typical
   interpretation would be something that also came before the  previous
   `v0.9.0` release.

</small>

If you were to re-order those commits by semantic versioning precedence, rather
than commit order, you would get:
```

|  git semver-describe    |  git describe            |
|-------------------------|--------------------------|
|                 v0.8.3  |       v0.8.3-1-g1a2b3c4* |
|      v0.8.3+1.g1a2b3c4  |       v0.8.3-2-g2b3c4d5* | 
|      v0.8.3+2.g2b3c4d5  |                  v0.8.3* |
|             v0.9.0-rc1  |   v0.9.0-rc1-1-gd3adb33* |
|  v0.9.0-rc1+1.gd3adb33  |              v0.9.0-rc1* | 
|                 v0.9.0  |                  v0.9.0  |

*: Out-of-order: SemVer ordering differs from commit history.
```


There are some other benefits:

- As a happy coincidence, the `+` character that precedes build metadata in
  SemVer v2.0 better expresses the situational context possible during a git
  describe: something that comes *after* a point in history, rather than
  *before* something to come.

## Related Work

- A similar versioning schematic was implemented as a NodeJS library, which
  looks to function by parsing and wrapping the output from shell commands sent
  to git: https://www.npmjs.com/package/git-describe.
- Some relevant discussion on the semver spec itself: semver/semver#106, semver/semver#106, semver/semver#106 