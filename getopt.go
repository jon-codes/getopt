// Package getopt provides a Go implementation of the Unix `getopt` function
// for parsing command-line options.
//
// This package parses of command-line arguments using the POSIX convention,
// supporting short options (e.g., -a) and arguments for these options. It also
// supports GNU extensions to allow for long options (e.g., --option) and
// options with optional arguments.
package getopt

import (
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

// type GetOptError string

// Errors that can be returned by GetOpt.
var (
	ErrDone          = errors.New("done")
	ErrUnknownOpt    = errors.New("unrecognized option")
	ErrIllegalOptArg = errors.New("option disallows arguments")
	ErrMissingOptArg = errors.New("option requires an argument")
)

type HasArg int

const (
	NoArgument       HasArg = iota // Indicates that the option disallows arguments.
	RequiredArgument               // Indicates that the option requires an argument.
	OptionalArgument               // Indicates that the option optionally accepts an argument.
)

type Func int

const (
	FuncGetOpt         Func = iota // Indicates that GetOpt should behave like `getopt`.
	FuncGetOptLong                 // Indicates that GetOpt should behave like `getopt_long`.
	FuncGetOptLongOnly             // Indicates that GetOpt should behave like `getopt_long_only`.
)

func (f Func) String() string {
	switch f {
	case FuncGetOpt:
		return "getopt"
	case FuncGetOptLong:
		return "getopt_long"
	case FuncGetOptLongOnly:
		return "getopt_long_only"
	default:
		return "unknown"
	}
}

type Mode int

const (
	ModeGNU Mode = iota
	ModePosix
	ModeInOrder
)

func (m Mode) String() string {
	switch m {
	case ModeGNU:
		return "gnu"
	case ModePosix:
		return "posix"
	case ModeInOrder:
		return "inorder"
	default:
		return "unknown"
	}
}

type Opt struct {
	Char   rune
	HasArg HasArg
}

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

type LongOpt struct {
	Name   string
	HasArg HasArg
}

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

type Params struct {
	Opts     []Opt
	LongOpts []LongOpt
	Func     Func
	Mode     Mode
}

type Result struct {
	Char   rune
	Name   string
	OptArg string
}

type State struct {
	args   []string // the current argument slice
	optInd int      // the next argument to process
	argInd int      // the next index of the current argument to process (when processing a short group)
}

const (
	initOptInd = 1
	initArgInd = 0
)

func NewState(args []string) *State {
	s := &State{
		args:   args,
		optInd: initOptInd,
		argInd: initArgInd,
	}
	return s
}

func (s *State) Args() []string {
	return s.args
}

func (s *State) OptInd() int {
	return s.optInd
}

func (s *State) Reset(args []string) {
	s.args = args
	s.optInd = initOptInd
	s.argInd = initArgInd
}

func (s *State) GetOpt(p Params) (res Result, err error) {
	if s.optInd >= len(s.args) {
		return res, ErrDone
	}

	if s.args[s.optInd] == "--" {
		s.optInd++
		return res, ErrDone
	}

	// TODO: cite permutation algo source (musl libc)
	pStart := s.optInd
	if s.args[s.optInd] == "-" || []rune(s.args[s.optInd])[0] != '-' {
		switch p.Mode {
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
		res, err = s.readOpt(p)
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

func (s *State) readOpt(p Params) (res Result, err error) {
	arg := s.args[s.optInd]
	checkLong := false
	if s.argInd == 0 {
		s.argInd++
		checkLong = p.Func == FuncGetOptLongOnly
		if arg[s.argInd] == '-' && p.Func != FuncGetOpt {
			s.argInd++
			checkLong = true
		}
	}

	hasArg := NoArgument
	name, inline, foundInline := strings.Cut(arg[s.argInd:], "=")

	if checkLong {
		overrideOpt := s.argInd == 1 && p.Func == FuncGetOptLongOnly
		opt, found := findLongOpt(name, overrideOpt, p)
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
		opt, found := findOpt(char, p)
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

func findOpt(char rune, p Params) (opt Opt, found bool) {
	i := slices.IndexFunc(p.Opts, func(s Opt) bool { return char == s.Char })
	if i >= 0 {
		return p.Opts[i], true
	} else {
		return opt, false
	}
}

func findLongOpt(name string, overrideOpt bool, p Params) (longOpt LongOpt, found bool) {
	if len([]rune(name)) == 1 && overrideOpt {
		_, found := findOpt([]rune(name)[0], p)
		if found {
			return longOpt, false
		}
	}

	matched := []LongOpt{}

	for _, lo := range p.LongOpts {
		if strings.HasPrefix(lo.Name, name) {
			matched = append(matched, lo)
		}
	}

	if len(matched) == 1 {
		return matched[0], true
	}

	return longOpt, false
}
