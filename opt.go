package opt

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

type GetOptErr string

const (
	ErrDone          = GetOptErr("done")
	ErrIllegalOpt    = GetOptErr("illegal option")
	ErrMissingOptArg = GetOptErr("missing required option argument")
)

func (e GetOptErr) Error() string {
	return string(e)
}

type HasArgType int

const (
	NoArgument HasArgType = iota
	RequiredArgument
	OptionalArgument
)

type Opt struct {
	Char   rune
	HasArg HasArgType
}

type LongOpt struct {
	Name   string
	HasArg HasArgType
}

type GetOptParams struct {
	Opts     []Opt
	LongOpts []LongOpt
}

type GetOptResult struct {
	Char   rune
	Name   string
	OptArg string
}

type GetOptState struct {
	Args     []string // the current argument slice
	OptIndex int      // the next argument to process
	ArgIndex int      // the next index of the current argument to process (when processing a short group)
}

func NewGetOptState(args []string) *GetOptState {
	g := &GetOptState{
		Args:     args,
		OptIndex: 1,
		ArgIndex: 0,
	}
	return g
}

func (g *GetOptState) Reset(args []string) {
	g.Args = args
	g.OptIndex = 1
	g.ArgIndex = 0
}

func (g *GetOptState) ReadOpt(p GetOptParams) (GetOptResult, error) {
	arg := g.Args[g.OptIndex]
	subArg := arg[g.ArgIndex:]
	r, _ := utf8.DecodeRuneInString(subArg) // TODO: handle errors

	if !isLegalOptRune(r) {
		return GetOptResult{Char: r}, ErrIllegalOpt
	}

	i := slices.IndexFunc(p.Opts, func(s Opt) bool { return r == s.Char })
	if i < 0 {
		return GetOptResult{Char: r}, ErrIllegalOpt
	}
	opt := p.Opts[i]
	g.ArgIndex++

	if arg[g.ArgIndex:] != "" {
		// there are more runes in the arg

		if opt.HasArg == RequiredArgument || opt.HasArg == OptionalArgument {
			// the rest of this arg is this options's argument
			optArg := arg[g.ArgIndex:]
			g.ArgIndex = 0
			g.OptIndex++
			return GetOptResult{Char: r, OptArg: optArg}, nil
		}

		return GetOptResult{Char: r}, nil
	} else {
		// this is the final rune in the arg
		g.ArgIndex = 0
		g.OptIndex++

		if opt.HasArg == NoArgument {
			return GetOptResult{Char: r}, nil
		}

		// look for an option argument in the next arg
		if g.OptIndex < len(g.Args) {
			nextArg := g.Args[g.OptIndex]
			g.OptIndex++
			// TODO: check for empty?
			return GetOptResult{Char: r, OptArg: nextArg}, nil
		}

		if opt.HasArg == RequiredArgument {
			return GetOptResult{Char: r}, ErrMissingOptArg
		}

		return GetOptResult{Char: r}, nil
	}
}

func (g *GetOptState) GetOpt(p GetOptParams) (GetOptResult, error) {
	if g.OptIndex >= len(g.Args) {
		// we've reached the end of the arg slice, so we're finished parsing.
		return GetOptResult{}, ErrDone
	}

	arg := g.Args[g.OptIndex]

	if g.ArgIndex == 0 {
		// we're attempting to parse a new argument
		if arg == "--" {
			// the '--' delimiter indicates that all remaining arguments are parameters
			g.OptIndex++
			return GetOptResult{}, ErrDone
		}
		if arg == "-" {
			// TODO: look into the expected behavior of a bare '-' option
			return GetOptResult{Char: '-'}, ErrIllegalOpt
		}
		if strings.HasPrefix(arg, "--") {
			// TODO: parse a long option
			return GetOptResult{Char: '-'}, ErrIllegalOpt
		}
		if strings.HasPrefix(arg, "-") {
			// parse a short option
			g.ArgIndex++
			return g.ReadOpt(p)
		}

		// this is a parameter, depending on the configuration, we either stop parsing, or permute it in the arg array
		should_permute := false // TODO: configure this dynamically
		if should_permute {
			// TODO: handle permute
			return GetOptResult{}, nil
		} else {
			// this and remaining args are parameters
			return GetOptResult{}, ErrDone
		}
	} else {
		// we're partway through parsing an option group (e.g. '-abc')
		return g.ReadOpt(p)
	}
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
