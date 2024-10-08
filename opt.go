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

type ShortOpt struct {
	Char   rune
	HasArg HasArgType
}

type LongOpt struct {
	Name   string
	HasArg HasArgType
}

type GetOptParams struct {
	Short []ShortOpt
	Long  []LongOpt
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

func (g *GetOptState) GetOpt(p GetOptParams) (GetOptResult, error) {
	if g.OptIndex >= len(g.Args) {
		return GetOptResult{}, ErrDone
	}

	arg := g.Args[g.OptIndex]

	if g.ArgIndex > 0 {
		// we're partway through parsing an option group
		r, _ := utf8.DecodeRuneInString(arg[g.ArgIndex:]) // TODO: error handling?
		if !isLegalOptRune(r) {
			return GetOptResult{Char: r}, ErrIllegalOpt
		}

		index := slices.IndexFunc(p.Short, func(s ShortOpt) bool { return r == s.Char })
		if index < 0 {
			return GetOptResult{Char: r}, ErrIllegalOpt
		}

		short := p.Short[index]
		g.ArgIndex++

		if arg[g.ArgIndex:] != "" {
			// there are more runes in the arg

			if short.HasArg == RequiredArgument || short.HasArg == OptionalArgument {
				// the rest of the arg is its argument
				optArg := arg[g.ArgIndex:]
				g.ArgIndex = 0
				g.OptIndex++
				return GetOptResult{Char: r, OptArg: optArg}, nil
			}

			return GetOptResult{Char: r}, nil
		} else {
			if short.HasArg == NoArgument {
				g.ArgIndex = 0
				g.OptIndex++
				return GetOptResult{Char: r}, nil
			}

			g.ArgIndex = 0
			g.OptIndex++

			if g.OptIndex < len(g.Args) {
				nextArg := g.Args[g.OptIndex]
				g.OptIndex++
				return GetOptResult{Char: r, OptArg: nextArg}, nil
				// }
			} else {
				if short.HasArg == RequiredArgument {
					return GetOptResult{}, ErrMissingOptArg
				} else {
					return GetOptResult{Char: r}, nil
				}
			}
		}
	} else {
		// we're parsing a new option
		if arg == "--" {
			g.OptIndex++
			return GetOptResult{}, ErrDone
		}

		if arg == "-" {
			return GetOptResult{Char: '-'}, ErrIllegalOpt
		}

		if strings.HasPrefix(arg, "--") {
			// TODO: we're parsing a long option
			return GetOptResult{Char: '-'}, ErrIllegalOpt
		}

		if strings.HasPrefix(arg, "-") {
			// we're parsing a short option
			g.ArgIndex = len("-")
			r, _ := utf8.DecodeRuneInString(arg[g.ArgIndex:]) // TODO: error handling?
			if !isLegalOptRune(r) {
				return GetOptResult{Char: r}, ErrIllegalOpt
			}

			index := slices.IndexFunc(p.Short, func(s ShortOpt) bool { return r == s.Char })
			if index > -1 {
				short := p.Short[index]
				g.ArgIndex++

				if arg[g.ArgIndex:] != "" {
					// there are more runes in the arg

					if short.HasArg == RequiredArgument || short.HasArg == OptionalArgument {
						// the rest of the arg is its argument
						optArg := arg[g.ArgIndex:]
						g.ArgIndex = 0
						g.OptIndex++
						return GetOptResult{Char: r, OptArg: optArg}, nil
					}

					return GetOptResult{Char: r}, nil
				} else {
					if short.HasArg == NoArgument {
						g.ArgIndex = 0
						g.OptIndex++
						return GetOptResult{Char: r}, nil
					}

					g.ArgIndex = 0
					g.OptIndex++

					if g.OptIndex < len(g.Args) {
						nextArg := g.Args[g.OptIndex]
						g.OptIndex++
						return GetOptResult{Char: r, OptArg: nextArg}, nil
						// }
					} else {
						if short.HasArg == RequiredArgument {
							return GetOptResult{}, ErrMissingOptArg
						} else {
							return GetOptResult{Char: r}, nil
						}
					}
				}
			} else {
				return GetOptResult{Char: r}, ErrIllegalOpt
			}
		} else {
			// this is a positional argument, so we're done parsing
			return GetOptResult{}, ErrDone
		}
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
