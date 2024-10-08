// Package getopt provides a Go implementation of the Unix `getopt` function
// for parsing command-line options.
//
// This package parses of command-line arguments using the POSIX convention,
// supporting short options (e.g., -a) and arguments for these options. It also
// supports GNU extensions to allow for long options (e.g., --option) and
// options with optional arguments.
package getopt

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

type GetOptErr string

// Errors that can be returned by GetOpt.
const (
	ErrDone          = GetOptErr("done")
	ErrUnknownOpt    = GetOptErr("unknown option")
	ErrMissingOptArg = GetOptErr("missing required option argument")
)

func (e GetOptErr) Error() string {
	return string(e)
}

type HasArg int

const (
	NoArgument       HasArg = iota // Indicates that the option disallows arguments.
	RequiredArgument               // Indicates that the option requires an argument.
	OptionalArgument               // Indicates that the option optionally accepts an argument.
)

type GetOptFunc int

const (
	FuncGetOpt         GetOptFunc = iota // Indicates that GetOpt should behave like `getopt`.
	FuncGetOptLong                       // Indicates that GetOpt should behave like `getopt_long`.
	FuncGetOptLongOnly                   // Indicates that GetOpt should behave like `getopt_long_only`.
)

type GetOptMode int

const (
	ModeGNU GetOptMode = iota
	ModePosix
	ModeInOrder
)

type Opt struct {
	Char   rune
	HasArg HasArg
}

func OptStr(optStr string) (opts []Opt) {
	for i := 0; i < len(optStr); i++ {
		char := rune(optStr[i])
		hasArg := NoArgument

		if i+1 < len(optStr) && optStr[i+1] == ':' {
			hasArg = RequiredArgument
			i++
			if i+1 < len(optStr) && optStr[i+1] == ':' {
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
	Function GetOptFunc
	Mode     GetOptMode
}

type Result struct {
	Char   rune
	Name   string
	OptArg string
}

type State struct {
	Args     []string // the current argument slice
	OptIndex int      // the next argument to process
	argIndex int      // the next index of the current argument to process (when processing a short group)
}

const (
	initOptIndex = 1
	initArgIndex = 0
)

func NewState(args []string) *State {
	s := &State{
		Args:     args,
		OptIndex: initOptIndex,
		argIndex: initArgIndex,
	}
	return s
}

func (s *State) Reset(args []string) {
	s.Args = args
	s.OptIndex = initOptIndex
	s.argIndex = initArgIndex
}

func (s *State) GetOpt(p Params) (res Result, err error) {
	if s.OptIndex >= len(s.Args) {
		return res, ErrDone
	}

	if s.Args[s.OptIndex] == "--" {
		s.OptIndex++
		return res, ErrDone
	}

	pStart := s.OptIndex
	if []rune(s.Args[s.OptIndex])[0] != '-' {
		switch p.Mode {
		case ModePosix:
			return res, ErrDone
		case ModeInOrder:
			s.OptIndex++
			return Result{Char: '\x01', OptArg: s.Args[s.OptIndex-1]}, nil
		default:
			for i := s.OptIndex; i < len(s.Args); i++ {
				arg := s.Args[i]
				if len(arg) > 1 && []rune(arg)[0] == '-' {
					s.OptIndex = i
					break
				}
				if i == len(s.Args)-1 {
					return res, ErrDone
				}
			}
		}
	}
	pEnd := s.OptIndex

	res, err = s.readOpt(p)

	if pEnd > pStart {
		count := s.OptIndex - pEnd
		for i := 0; i < count; i++ {
			s.permute(s.OptIndex-1, pStart)
		}
		s.OptIndex = pStart + count
	}
	return res, err
}

func (s *State) permute(src, dest int) {
	tmp := s.Args[src]
	for i := src; i > dest; i-- {
		s.Args[i] = s.Args[i-1]
	}
	s.Args[dest] = tmp
}

func findOpt(char rune, p Params) (opt Opt, found bool) {
	if !isLegalOptRune(char) {
		return opt, false
	}
	i := slices.IndexFunc(p.Opts, func(s Opt) bool { return char == s.Char })
	if i >= 0 {
		return p.Opts[i], true
	} else {
		return opt, false
	}
}

func findLongOpt(name string, p Params) (longOpt LongOpt, found bool) {
	for _, r := range name {
		if !isLegalOptRune(r) {
			return longOpt, false
		}
	}

	matched := []LongOpt{}

	for _, lo := range p.LongOpts {
		if len([]rune(name)) == 1 {
			i := slices.IndexFunc(p.Opts, func(s Opt) bool { return []rune(name)[0] == s.Char })
			if i >= 0 {
				return lo, true
			}
		}
		if strings.HasPrefix(lo.Name, name) {
			matched = append(matched, lo)
		}
	}

	if len(matched) == 1 {
		return matched[0], true
	}

	return longOpt, false
}

func (s *State) readOpt(p Params) (res Result, err error) {
	arg := s.Args[s.OptIndex]
	checkLong := false
	if s.argIndex == 0 {
		s.argIndex++
		checkLong = p.Function == FuncGetOptLongOnly
		if arg[s.argIndex] == '-' && p.Function != FuncGetOpt {
			s.argIndex++
			checkLong = true
		}
	}

	hasArg := NoArgument
	checkNext := true

	if checkLong {
		name, inline, foundInline := strings.Cut(arg[s.argIndex:], "=")
		opt, found := findLongOpt(arg[s.argIndex:], p)
		if found {
			s.OptIndex++
			hasArg = opt.HasArg
			res.Name = name
			if foundInline {
				res.OptArg = inline
				checkNext = false
			}
			s.argIndex = 0
		}
	}

	if res.Name == "" {
		char, _ := utf8.DecodeRuneInString(arg[s.argIndex:])
		res.Char = char
		opt, found := findOpt(char, p)
		if found {
			s.argIndex++
			hasArg = opt.HasArg

			if arg[s.argIndex:] == "" {
				s.OptIndex++
				s.argIndex = 0
			} else if hasArg != NoArgument {
				res.OptArg = arg[s.argIndex:]
				s.argIndex = 0
				checkNext = false
				s.OptIndex++
			}
		} else {
			s.OptIndex++
			err = ErrUnknownOpt
		}
	}

	if checkNext && hasArg != NoArgument && s.OptIndex < len(s.Args) {
		if s.Args[s.OptIndex] == "--" {
			s.OptIndex++
		} else {
			res.OptArg = s.Args[s.OptIndex]
			s.OptIndex++
		}
	}

	if res.Char == 0 && res.Name == "" {
		err = ErrUnknownOpt
	}

	if res.OptArg != "" && hasArg == NoArgument {
		err = ErrUnknownOpt
	}

	if res.OptArg == "" && hasArg == RequiredArgument {
		err = ErrMissingOptArg
	}

	return res, err
}

func isGraph(r rune) bool {
	// POSIX 7.3.1
	// > Define characters to be classified as punctuation characters.
	// > In the POSIX locale, neither the <space> nor any characters in classes alpha, digit, or cntrl shall be included.
	return unicode.IsDigit(r) || unicode.IsLetter(r) || unicode.IsPunct(r)
}

func isLegalOptRune(r rune) bool {
	// > A legitimate option character is any visible one byte ascii(7)
	// > character (for which isgraph(3) would return nonzero) that is not '-', ':', or ';'.)
	return r != ':' && r != ';' && r <= unicode.MaxASCII && isGraph(r)
}
