# getopt
[![Go Reference](https://pkg.go.dev/badge/github.com/jon-codes/getopt.svg)](https://pkg.go.dev/github.com/jon-codes/getopt)
[![CI](https://github.com/jon-codes/getopt/actions/workflows/ci.yml/badge.svg)](https://github.com/jon-codes/getopt/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jon-codes/getopt)](https://goreportcard.com/report/github.com/jon-codes/getopt)
[![codecov](https://codecov.io/github/jon-codes/getopt/graph/badge.svg?token=CF7WDJOFVY)](https://codecov.io/github/jon-codes/getopt)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Package `getopt` provides a zero-dependency Go implementation of the Unix getopt function for parsing command-line options.

The `getopt` package supports parsing options using the POSIX convention, supporting short options (e.g., `-a`) and option arguments. It also supports GNU extensions, including support for long options (e.g., `--option`), options with optional arguments, and permuting non-option parameters. 

## Install

```
go get github.com/jon-codes/getopt
```

## Usage

This package emulates the C `getopt` function, but uses a state machine to encapsulate variables (instead of the global `optind`, `optopt`, `optarg` used in C). Rather than implement a high-level interface for defining CLI flags, it aims to implement an accurate emulation of C `getopt` that can be used by higher-level tools.

Collect all options into a slice:

```go
state := getopt.NewState(os.Args)
config := getopt.Config{Opts: getopt.OptStr(`ab:c::`)}
opts, err := state.Parse(config)
```

Iterate over each option for finer control:

```go
state := getopt.NewState(os.Args)
config := getopt.Config{Opts: getopt.OptStr(`ab:c::`)}

for opt, err := range state.All(config) {
    if err != nil {
        break
    }
    switch opt.Char {
    case 'a':
        fmt.Printf("Found opt a\n")
    case 'b':
        fmt.Printf("Found opt b with arg %s\n", opt.OptArg)
    case 'c':
        fmt.Printf("Found opt c")
        if opt.OptArg != "" {
            fmt.Printf(" with arg %s", opt.OptArg)
        }
        fmt.Printf("\n")
    }
}
```
# Behavior

This package uses [GNU libc](https://www.gnu.org/software/libc/) as a reference for behavior, since many expect the
non-standard features it provides. This is accomplished via a C test generator that runs getopt for all functions and parsing modes.

It supports the same configuration options as the GNU options via [Mode](https://pkg.go.dev/github.com/jon-codes/getopt#Mode):
  - [ModeGNU](https://pkg.go.dev/github.com/jon-codes/getopt#ModeGNU): enables default behavior.
  - [ModePosix](https://pkg.go.dev/github.com/jon-codes/getopt#ModePosix): enables the '+' compatibility mode, disabling permuting arguments and terminating parsing on the first parameter.
  - [ModeInOrder](https://pkg.go.dev/github.com/jon-codes/getopt#ModeInOrder): enables the '-' optstring prefix mode, treating all parameters as though they were arguments to an option with character code 1.

The specific libc function that is emulated can be configured via [Func](https://pkg.go.dev/github.com/jon-codes/getopt#Func):
  - [FuncGetOpt](https://pkg.go.dev/github.com/jon-codes/getopt#FuncGetOpt): parse only traditional POSIX short options (e.g., -a).
  - [FuncGetOptLong](https://pkg.go.dev/github.com/jon-codes/getopt#FuncGetOptLong): parse short options, and GNU extension long options (e.g.,
    --option).
  - [FuncGetOptLongOnly](https://pkg.go.dev/github.com/jon-codes/getopt#FuncGetOptLongOnly): parse short and long options, but allow long options to begin with a single dash (like [pkg/flag](https://pkg.go.dev/flag)).

The parser differs from GNU libc's getopt in the following ways:
  - It accepts multi-byte runes in short and long option definitions.
  - This package does not implement the same argument permutation as GNU libc.
    The value of OptInd and order of arguments mid-parsing may differ, and only
    the final order is validated against the GNU implementation.

## API Documentation

The full API documentation can be found at [pkg.go.dev](https://pkg.go.dev/github.com/jon-codes/getopt).

## Acknowledgements

The algorithm for permuting arguments is from [musl-libc](https://git.musl-libc.org/cgit/musl/tree/COPYRIGHT), and is used under the linked MIT License:

 | Copyright © 2005-2020 Rich Felker, et al.
