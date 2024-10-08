/*
Package getopt implements the Unix [getopt] function for parsing command-line
options.

[getopt]: https://www.man7.org/linux/man-pages/man3/getopt.3.html
*/
package getopt

import (
	"errors"
	"iter"
	"slices"
	"strings"
	"unicode/utf8"
)

// Errors that can be returned during option parsing.
var (
	ErrDone          = errors.New("getopt: done")
	ErrUnknownOpt    = errors.New("getopt: unrecognized option")
	ErrIllegalOptArg = errors.New("getopt: option disallows arguments")
	ErrMissingOptArg = errors.New("getopt: option requires an argument")
)

// HasArg defines rules for parsing option arguments.
type HasArg int

const (
	NoArgument       HasArg = iota // option may not take an argument
	RequiredArgument               // option requires an argument
	OptionalArgument               // option may optionally accept an argument
)

// Func indicates which POSIX or GNU extension function to emulate during option
// parsing.
type Func int

const (
	FuncGetOpt         Func = iota // behave like `getopt`
	FuncGetOptLong                 // behave like `getopt_long`
	FuncGetOptLongOnly             // behave like `getopt_long_only`
)

// Mode indicates which behavior to enable during option parsing.
type Mode int

const (
	ModeGNU     Mode = iota // enable GNU extension behavior
	ModePosix               // enable POSIX behavior (terminate on first parameter)
	ModeInOrder             // enable "in-order" behavior (parse parameters as options)
)

// An Opt is a parsing rule for a short, single-character command-line option
// (e.g., -a).
type Opt struct {
	Char   rune   // option character
	HasArg HasArg // option argument rule
}

// OptStr parses an option string, returning a slice of Opt.
//
// The option string uses the same format as [getopt], with each character
// representing an option. Options with a single ":" suffix require an argument.
// Options with a double "::" suffix optionally accept an argument. Options with
// no suffix do not allow arguments.
//
// [getopt]: https://www.man7.org/linux/man-pages/man3/getopt.3.html
func OptStr(optStr string) (opts []Opt) {
	var i int
	for i < len(optStr) {
		char, size := utf8.DecodeRuneInString(optStr[i:])
		i += size

		hasArg := NoArgument
		if i < len(optStr) && optStr[i] == ':' {
			hasArg = RequiredArgument
			i++
			if i < len(optStr) && optStr[i] == ':' {
				hasArg = OptionalArgument
				i++
			}
		}
		opts = append(opts, Opt{Char: char, HasArg: hasArg})
	}

	return opts
}

// A LongOpt is a parsing rule for a named, long command-line option
// (e.g., --option).
type LongOpt struct {
	Name   string // option name
	HasArg HasArg // option argument rule
}

// OptStr parses a long option string, returning a slice of LongOpt.
//
// The option string uses the same format as --longoptions in the GNU
// [getopt(1)] command. Option names are comma-separated, and argument rules are
// designated by colon suffixes, like with [OptStr].
//
// [getopt(1)]: https://www.man7.org/linux/man-pages/man1/getopt.1.html
func LongOptStr(longOptStr string) (longOpts []LongOpt) {
	items := strings.Split(longOptStr, ",")
	if len(items) == 1 && items[0] == "" {
		return longOpts
	}
	for _, item := range items {
		var opt LongOpt
		opt.Name = strings.TrimRight(item, ":")
		if len(opt.Name) == len(item)-1 {
			opt.HasArg = RequiredArgument
		} else if len(opt.Name) == len(item)-2 {
			opt.HasArg = OptionalArgument
		}

		longOpts = append(longOpts, opt)
	}

	return longOpts
}

// A Config defines the rules and behavior used when parsing options. Note the
// zero values for Func ([FuncGetOpt]) and Mode ([ModeGNU]), which will
// determine the parsing behavior unless set otherwise.
type Config struct {
	Opts     []Opt     // allowed short options
	LongOpts []LongOpt // allowed long options
	Func     Func      // parsing function
	Mode     Mode      // parsing behavior
}

type Result struct {
	Char   rune   // parsed short option character
	Name   string // parsed long option name
	OptArg string // parsed option argument
}

type State struct {
	args   []string // current argument slice
	optInd int      // next argument to process
	argInd int      // next index of the current argument to process (when processing a short group)
}

const (
	initOptInd = 1
	initArgInd = 0
)

// NewState returns a new [State] to parse options from args, starting with the
// element at index 1.
func NewState(args []string) *State {
	s := &State{
		args:   args,
		optInd: initOptInd,
		argInd: initArgInd,
	}
	return s
}

// Parse returns a slice of [Result] by calling [State.GetOpt] until an error is
// returned.
func (s *State) Parse(c Config) ([]Result, error) {
	results := []Result{}
	for res, err := range s.All(c) {
		if err != nil {
			if err == ErrDone {
				return results, nil
			}
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}

// All returns an iterator that yields successive parsing results.
func (s *State) All(c Config) iter.Seq2[Result, error] {
	return func(yield func(Result, error) bool) {
		for {
			result, err := s.GetOpt(c)
			if err == ErrDone {
				return
			}
			if !yield(result, err) {
				return
			}
		}
	}
}

// Args returns the current slice of arguments in [State].
// This may differ from the slice used to initialize State, since parsing can
// permute the argument order.
func (s *State) Args() []string {
	return s.args
}

// OptInd returns the index of the next argument that will be parsed in State.
//
// After all options have been parsed, OptInd will index the first parameter
// (non-option) argument returned by [State.Args]. If no parameters are present,
// the index will be invalid, since the next argument's index would have
// exceeded the bounds of the argument slice. [State.Params] provides safe
// access to parameters.
func (s *State) OptInd() int {
	return s.optInd
}

// Params returns the slice of parameter (non-option) arguments. If parsing has
// not successfully completed with [ErrDone], this may include arguments that
// otherwise be parsed as options.
func (s *State) Params() []string {
	if s.optInd >= len(s.args) {
		return []string{}
	}
	return s.args[s.optInd:]
}

// Reset recycles an existing [State], resetting it to parse options from args,
// starting with the element at index 1.
func (s *State) Reset(args []string) {
	s.args = args
	s.optInd = initOptInd
	s.argInd = initArgInd
}

// GetOpt returns the result of parsing the next option in [State].
//
// If parsing has successfully completed, err will be [ErrDone]. Otherwise, the
// returned [Result] indicates either a valid option, or the properties of an
// invalid option if err is non-nil.
func (s *State) GetOpt(c Config) (res Result, err error) {
	if s.optInd >= len(s.args) {
		return res, ErrDone
	}

	if s.args[s.optInd] == "--" {
		s.optInd++
		return res, ErrDone
	}

	// The algorithm for permuting arguments is from [musl-libc], and is used under the MIT License:
	// Copyright © 2005-2020 Rich Felker, et al.
	pStart := s.optInd
	if s.args[s.optInd] == "" || s.args[s.optInd] == "-" || []rune(s.args[s.optInd])[0] != '-' {
		switch c.Mode {
		case ModePosix:
			return res, ErrDone
		case ModeInOrder:
			s.optInd++
			return Result{Char: '\x01', OptArg: s.args[s.optInd-1]}, nil
		default:
			for i := s.optInd; i < len(s.args); i++ {
				arg := s.args[i]
				if len(arg) > 1 && []rune(arg)[0] == '-' {
					s.optInd = i
					break
				}
				if i == len(s.args)-1 {
					return res, ErrDone
				}
			}
		}
	}
	pEnd := s.optInd

	if s.args[s.optInd] == "--" {
		s.optInd++
		err = ErrDone
	} else {
		res, err = s.readOpt(c)
	}

	if pEnd > pStart {
		count := s.optInd - pEnd
		for i := 0; i < count; i++ {
			s.permute(s.optInd-1, pStart)
		}
		s.optInd = pStart + count
	}
	return res, err
}

func (s *State) readOpt(c Config) (res Result, err error) {
	arg := s.args[s.optInd]
	checkLong := false
	if s.argInd == 0 {
		s.argInd++
		checkLong = c.Func == FuncGetOptLongOnly
		if arg[s.argInd] == '-' && c.Func != FuncGetOpt {
			s.argInd++
			checkLong = true
		}
	}

	hasArg := NoArgument
	name, inline, foundInline := strings.Cut(arg[s.argInd:], "=")

	if checkLong && name != "" {
		overrideOpt := s.argInd == 1 && c.Func == FuncGetOptLongOnly
		opt, found := findLongOpt(name, overrideOpt, c)
		if found {
			s.optInd++
			hasArg = opt.HasArg
			res.Name = opt.Name
			if foundInline {
				if opt.HasArg == NoArgument {
					err = ErrIllegalOptArg
				}
				res.OptArg = inline
			}
			s.argInd = 0
		}
	}

	if res.Name == "" {
		char, size := utf8.DecodeRuneInString(arg[s.argInd:])
		res.Char = char
		opt, found := findOpt(char, c)
		if found {
			s.argInd += size
			hasArg = opt.HasArg

			if arg[s.argInd:] == "" {
				s.optInd++
				s.argInd = 0
			} else if hasArg != NoArgument {
				res.OptArg = arg[s.argInd:]
				s.argInd = 0
				s.optInd++
			}
		} else {
			s.argInd++
			if checkLong {
				s.optInd++
				s.argInd = 0
				res.Char = 0
				res.Name = name
			} else if arg[s.argInd:] == "" {
				s.optInd++
				s.argInd = 0
			}
			err = ErrUnknownOpt
		}
	}

	if hasArg == RequiredArgument && res.OptArg == "" && s.optInd < len(s.args) {
		res.OptArg = s.args[s.optInd]
		s.optInd++
	}

	if res.OptArg != "" && hasArg == NoArgument {
		err = ErrIllegalOptArg
	}

	if res.OptArg == "" && hasArg == RequiredArgument {
		err = ErrMissingOptArg
	}

	return res, err
}

func (s *State) permute(src, dest int) {
	tmp := s.args[src]
	for i := src; i > dest; i-- {
		s.args[i] = s.args[i-1]
	}
	s.args[dest] = tmp
}

func findOpt(char rune, c Config) (opt Opt, found bool) {
	i := slices.IndexFunc(c.Opts, func(s Opt) bool { return char == s.Char })
	if i >= 0 {
		return c.Opts[i], true
	} else {
		return opt, false
	}
}

func findLongOpt(name string, overrideOpt bool, c Config) (longOpt LongOpt, found bool) {
	if len([]rune(name)) == 1 && overrideOpt {
		_, found := findOpt([]rune(name)[0], c)
		if found {
			return longOpt, false
		}
	}

	matched := []LongOpt{}

	for _, lo := range c.LongOpts {
		if strings.HasPrefix(lo.Name, name) {
			matched = append(matched, lo)
		}
	}

	if len(matched) == 1 {
		return matched[0], true
	}

	return longOpt, false
}
