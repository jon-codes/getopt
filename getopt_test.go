package getopt

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"unicode"
)

func TestFindOpt(t *testing.T) {
	t.Run("it finds opts", func(t *testing.T) {
		p := Params{Opts: OptStr(`ab`)}
		got, found := findOpt('b', p)
		want := 'b'

		if !found {
			t.Fatalf("didn't find an option, but wanted to")
		}
		if got.Char != want {
			t.Errorf("got Char %q, but wanted %q", got.Char, want)
		}
	})

	t.Run("it finds multi-byte opts", func(t *testing.T) {
		p := Params{Opts: OptStr(`αβ`)}
		got, found := findOpt('β', p)
		want := 'β'

		if !found {
			t.Fatalf("didn't find an option, but wanted to")
		}
		if got.Char != want {
			t.Errorf("got Char %q, but wanted %q", got.Char, want)
		}
	})

	t.Run("it doesn't find nonexistent opts", func(t *testing.T) {
		p := Params{Opts: OptStr(`ab`)}
		got, found := findOpt('c', p)

		if found {
			t.Errorf("found option with Char %q, but didn't want to", got.Char)
		}
	})
}

func TestFindLongOpt(t *testing.T) {
	t.Run("it finds exact long opts", func(t *testing.T) {
		p := Params{LongOpts: LongOptStr(`longa,longb`)}
		got, found := findLongOpt("longb", false, p)
		want := "longb"

		if !found {
			t.Fatalf("didn't find an option, but wanted to")
		}
		if got.Name != want {
			t.Errorf("got Name %q, but wanted %q", got.Name, want)
		}
	})

	t.Run("it finds exact multi-byte long opts", func(t *testing.T) {
		p := Params{LongOpts: LongOptStr(`longa,αβγ`)}
		got, found := findLongOpt("αβγ", false, p)
		want := "αβγ"

		if !found {
			t.Fatalf("didn't find an option, but wanted to")
		}
		if got.Name != want {
			t.Errorf("got Name %q, but wanted %q", got.Name, want)
		}
	})

	t.Run("it finds uniq abbreviated long opts", func(t *testing.T) {
		p := Params{LongOpts: LongOptStr(`longa_blah,longb_blah`)}
		got, found := findLongOpt("longb", false, p)
		want := "longb_blah"

		if !found {
			t.Fatalf("didn't find an option, but wanted to")
		}
		if got.Name != want {
			t.Errorf("got Name %q, but wanted %q", got.Name, want)
		}
	})

	t.Run("it finds uniq abbreviated multi-byte long opts", func(t *testing.T) {
		p := Params{LongOpts: LongOptStr(`αβεδ,αβγδ`)}
		got, found := findLongOpt("αβε", false, p)
		want := "αβεδ"

		if !found {
			t.Fatalf("didn't find an option, but wanted to")
		}
		if got.Name != want {
			t.Errorf("got Name %q, but wanted %q", got.Name, want)
		}
	})

	t.Run("it doesn't find nonexistent long opts", func(t *testing.T) {
		p := Params{LongOpts: LongOptStr(`longa,longb`)}
		got, found := findLongOpt("longc", false, p)

		if found {
			t.Errorf("found option with Name %q, but didn't want to", got.Name)
		}
	})

	t.Run("it does not find non-uniq abbreviated long opts", func(t *testing.T) {
		p := Params{LongOpts: LongOptStr(`long,longa,longb`)}
		got, found := findLongOpt("long", false, p)

		if found {
			t.Errorf("found option with Name %q, but didn't want to", got.Name)
		}
	})

	t.Run("without overrideOpt, it defers to matching opts", func(t *testing.T) {
		p := Params{Opts: OptStr(`l`), LongOpts: LongOptStr(`longa`)}
		got, found := findLongOpt("l", true, p)

		if found {
			t.Errorf("found option with Name %q, but didn't want to", got.Name)
		}
	})

	t.Run("without overrideOpt, it defers to matching multi-byte opts", func(t *testing.T) {
		p := Params{Opts: OptStr(`α`), LongOpts: LongOptStr(`αβγ`)}
		got, found := findLongOpt("α", true, p)

		if found {
			t.Errorf("found option with Name %q, but didn't want to", got.Name)
		}
	})

	t.Run("with overrideOpt it allows abbreviations matching opts", func(t *testing.T) {
		p := Params{Opts: OptStr(`l`), LongOpts: LongOptStr(`longa`)}
		got, found := findLongOpt("l", false, p)
		want := "longa"

		if !found {
			t.Fatalf("didn't find an option, but wanted to")
		}
		if got.Name != want {
			t.Errorf("got Name %q, but wanted %q", got.Name, want)
		}
	})
}

func TestOptStr(t *testing.T) {
	t.Run("it handles empty values", func(t *testing.T) {
		got := OptStr(``)

		if len(got) != 0 {
			t.Errorf("got length %d, but wanted 0", len(got))
		}
	})

	t.Run("it parses opts", func(t *testing.T) {
		got := OptStr(`ab:c::d:e`)
		want := []Opt{
			{Char: 'a', HasArg: NoArgument},
			{Char: 'b', HasArg: RequiredArgument},
			{Char: 'c', HasArg: OptionalArgument},
			{Char: 'd', HasArg: RequiredArgument},
			{Char: 'e', HasArg: NoArgument},
		}

		if !slices.Equal(got, want) {
			t.Errorf("got %+v, but wanted %+v", got, want)
		}
	})

	t.Run("it parses multi-byte opts", func(t *testing.T) {
		got := OptStr(`αβ:γ::δ:ε`)
		want := []Opt{
			{Char: 'α', HasArg: NoArgument},
			{Char: 'β', HasArg: RequiredArgument},
			{Char: 'γ', HasArg: OptionalArgument},
			{Char: 'δ', HasArg: RequiredArgument},
			{Char: 'ε', HasArg: NoArgument},
		}

		if !slices.Equal(got, want) {
			t.Errorf("got %+v, but wanted %+v", got, want)
		}
	})
}

func TestLongOptStr(t *testing.T) {
	t.Run("it handles empty values", func(t *testing.T) {
		got := LongOptStr(``)

		if len(got) != 0 {
			t.Errorf("got length %d, but wanted 0", len(got))
		}
	})

	t.Run("it parses long opts", func(t *testing.T) {
		got := LongOptStr(`long_a,long_b:,long_c::`)
		want := []LongOpt{
			{Name: "long_a", HasArg: NoArgument},
			{Name: "long_b", HasArg: RequiredArgument},
			{Name: "long_c", HasArg: OptionalArgument},
		}

		if !slices.Equal(got, want) {
			t.Errorf("got %+v, but wanted %+v", got, want)
		}
	})

	t.Run("it parses long opts with multi-byte elements", func(t *testing.T) {
		got := LongOptStr(`άλφα,βήτα:,γάμμα::`)
		want := []LongOpt{
			{Name: "άλφα", HasArg: NoArgument},
			{Name: "βήτα", HasArg: RequiredArgument},
			{Name: "γάμμα", HasArg: OptionalArgument},
		}

		if !slices.Equal(got, want) {
			t.Errorf("got %+v, but wanted %+v", got, want)
		}
	})
}

func TestNewState(t *testing.T) {
	got := NewState(argsStr(`prgm -a -b`))
	want := State{OptIndex: 1, Args: []string{"prgm", "-a", "-b"}}

	if got.OptIndex != want.OptIndex {
		t.Errorf("got %d, but wanted %d", got.OptIndex, want.OptIndex)
	}

	if !slices.Equal(got.Args, want.Args) {
		t.Errorf("got %+q, but wanted %+q", got.Args, want.Args)
	}
}

func TestStateReset(t *testing.T) {
	s := NewState(argsStr(`prgm -a -b`))
	p := Params{Opts: OptStr(`abc`)}

	assertGetOpt(t, s, p, assertion{
		char:     'a',
		args:     argsStr(`prgm -a -b`),
		optIndex: 2,
	})

	s.Reset(argsStr(`prgm -c -b`))

	assertGetOpt(t, s, p, assertion{
		char:     'c',
		args:     argsStr(`prgm -c -b`),
		optIndex: 2,
	})
}

func TestStatePermute(t *testing.T) {
	t.Run("it moves an arg backwards from source to dest", func(t *testing.T) {
		s := NewState(argsStr(`prgm a b c d`))
		s.permute(3, 1)
		want := argsStr(`prgm c a b d`)

		if !slices.Equal(s.Args, want) {
			t.Errorf("got Args %q, but wanted %q", s.Args, want)
		}
	})
}

func TestStateGetOpt_FuncGetOpt(t *testing.T) {
	function := FuncGetOpt
	mode := ModeGNU

	t.Run("it handles opts with no arguments", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a -bc`))
		p := Params{Opts: OptStr(`abc`), Function: function, Mode: mode}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -bc`), optIndex: 2}, // bare opt
			{char: 'b', args: argsStr(`prgm -a -bc`), optIndex: 2}, // first opt in group
			{char: 'c', args: argsStr(`prgm -a -bc`), optIndex: 3}, // second opt in group
			{err: ErrDone, args: argsStr(`prgm -a -bc`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it handles opts with required arguments", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a -b -bc -c -- -d`))
		p := Params{Opts: OptStr(`a:b:c:d:`), Function: function, Mode: mode}

		wants := []assertion{
			{char: 'a', optArg: "-b", args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 3},          // bare opt, consumes next arg
			{char: 'b', optArg: "c", args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 4},           // first opt in group, consumes group
			{char: 'c', optArg: "--", args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 6},          // bare opt, consume '--' as arg
			{char: 'd', err: ErrMissingOptArg, args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 7}, // bare opt, no next arg
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it handles opts with optional arguments", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a -b -bc -c -- -d`))
		p := Params{Opts: OptStr(`a::b::c::d::`), Function: function, Mode: mode}

		wants := []assertion{
			{char: 'a', optArg: "-b", args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 3}, // bare opt, consumes next arg
			{char: 'b', optArg: "c", args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 4},  // first opt in group, consumes group
			{char: 'c', optArg: "--", args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 6}, // bare opt, consume '--' as arg
			{char: 'd', args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 7},               // bare opt, no next arg
			{err: ErrDone, args: argsStr(`prgm -a -b -bc -c -- -d`), optIndex: 7},
		}

		assertSeq(t, s, p, wants)
	})
}

func TestStateGetOpt_FuncGetOptLong(t *testing.T) {
	function := FuncGetOptLong
	mode := ModeGNU

	t.Run("it handles long opts with no arguments", func(t *testing.T) {
		s := NewState(argsStr(`prgm --longa --longb p1 --longc=`))
		p := Params{LongOpts: LongOptStr(`longa,longb,longc`), Function: function, Mode: mode}

		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa --longb p1 --longc=`), optIndex: 2},
			{name: "longb", args: argsStr(`prgm --longa --longb p1 --longc=`), optIndex: 3},                        // treat next arg as param
			{name: "longc", err: ErrIllegalOptArg, args: argsStr(`prgm --longa --longb --longc= p1`), optIndex: 4}, // disallow args
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it handles long opts with required arguments", func(t *testing.T) {
		s := NewState(argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`))
		p := Params{LongOpts: LongOptStr(`longa:,longb:,longc:,longd:`), Function: function, Mode: mode}

		wants := []assertion{
			{name: "longa", optArg: "--longb", args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 3},     // bare opt, consumes next arg
			{name: "longb", optArg: "arg1", args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 4},        // inline opt arg
			{name: "longc", optArg: "--", args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 6},          // bare opt, consume '--' as arg
			{name: "longd", err: ErrMissingOptArg, args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 7}, // bare opt, no next arg
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it handles long opts with optional arguments", func(t *testing.T) {
		s := NewState(argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`))
		p := Params{LongOpts: LongOptStr(`longa::,longb::,longc::,longd::`), Function: function, Mode: mode}

		wants := []assertion{
			{name: "longa", optArg: "--longb", args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 3}, // bare opt, consumes next arg
			{name: "longb", optArg: "arg1", args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 4},    // first opt in group, consumes group
			{name: "longc", optArg: "--", args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 6},      // bare opt, consume '--' as arg
			{name: "longd", args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 7},                    // bare opt, no next arg
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longb=arg1 --longc -- --longd`), optIndex: 7},
		}

		assertSeq(t, s, p, wants)
	})
}

func TestGetOpt_FuncGetOpt(t *testing.T) {
	function := FuncGetOpt

	t.Run("it parses opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -abc -d p1`))
		p := Params{Opts: OptStr(`abcd`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -abc -d p1`), optIndex: 1},
			{char: 'b', args: argsStr(`prgm -abc -d p1`), optIndex: 1},
			{char: 'c', args: argsStr(`prgm -abc -d p1`), optIndex: 2},
			{char: 'd', args: argsStr(`prgm -abc -d p1`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -abc -d p1`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses multi-byte opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -αβγ -δ p1`))
		p := Params{Opts: OptStr(`αβγδ`), Function: function}

		wants := []assertion{
			{char: 'α', args: argsStr(`prgm -αβγ -δ p1`), optIndex: 1},
			{char: 'β', args: argsStr(`prgm -αβγ -δ p1`), optIndex: 1},
			{char: 'γ', args: argsStr(`prgm -αβγ -δ p1`), optIndex: 2},
			{char: 'δ', args: argsStr(`prgm -αβγ -δ p1`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -αβγ -δ p1`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it errors on undefined opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a`))
		p := Params{Opts: OptStr(`b`), Function: function}

		wants := []assertion{
			{char: 'a', err: ErrUnknownOpt, args: argsStr(`prgm -a`), optIndex: 1},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses multiple opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a -b p1`))
		p := Params{Opts: OptStr(`ab`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -b p1`), optIndex: 2},
			{char: 'b', args: argsStr(`prgm -a -b p1`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -a -b p1`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses opt groups", func(t *testing.T) {
		s := NewState(argsStr(`prgm -ab p1`))
		p := Params{Opts: OptStr(`ab`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -ab p1`), optIndex: 1},
			{char: 'b', args: argsStr(`prgm -ab p1`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -ab p1`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses required opt args in the next argument", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a arg_a`))
		p := Params{Opts: OptStr(`a:`), Function: function}

		wants := []assertion{
			{char: 'a', optArg: "arg_a", args: argsStr(`prgm -a arg_a`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -a arg_a`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses required opt args in same argument", func(t *testing.T) {
		s := NewState(argsStr(`prgm -aarg_a`))
		p := Params{Opts: OptStr(`a:`), Function: function}

		wants := []assertion{
			{char: 'a', optArg: "arg_a", args: argsStr(`prgm -aarg_a`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -aarg_a`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it errors on missing required opt args", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a`))
		p := Params{Opts: OptStr(`a:`), Function: function}

		wants := []assertion{
			{char: 'a', err: ErrMissingOptArg, args: argsStr(`prgm -a`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses opts with missing optional opt args", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a`))
		p := Params{Opts: OptStr(`a::`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -a`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses optional opt args in the next argument", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a arg_a`))
		p := Params{Opts: OptStr(`a::`), Function: function}

		wants := []assertion{
			{char: 'a', optArg: "arg_a", args: argsStr(`prgm -a arg_a`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -a arg_a`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses optional opt args in same argument", func(t *testing.T) {
		s := NewState(argsStr(`prgm -aarg_a`))
		p := Params{Opts: OptStr(`a::`), Function: function}

		wants := []assertion{
			{char: 'a', optArg: "arg_a", args: argsStr(`prgm -aarg_a`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -aarg_a`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it treats arguments after '--' as parameters", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a -- -b`))
		p := Params{Opts: OptStr(`ab`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -- -b`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -a -- -b`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it allows '--' as a potential option argument", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a -- p1`))
		p := Params{Opts: OptStr(`a::`), Function: function}

		wants := []assertion{
			{char: 'a', optArg: "--", args: argsStr(`prgm -a -- p1`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -a -- p1`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it skips and permutes non-opt params in gnu mode", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a pos1 -b`))
		p := Params{Opts: OptStr(`ab`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a pos1 -b`), optIndex: 2},
			{char: 'b', args: argsStr(`prgm -a -b pos1`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -a -b pos1`), optIndex: 3},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it terminates on non-opt params in posix mode", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a pos1 -b`))
		p := Params{Opts: OptStr(`ab`), Function: function, Mode: ModePosix}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a pos1 -b`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -a pos1 -b`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it treats non-opt params as opts in in-order mode", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a pos1 -b`))
		p := Params{Opts: OptStr(`ab`), Function: function, Mode: ModeInOrder}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a pos1 -b`), optIndex: 2},
			{char: '\x01', optArg: "pos1", args: argsStr(`prgm -a pos1 -b`), optIndex: 3},
			{char: 'b', args: argsStr(`prgm -a pos1 -b`), optIndex: 4},
			{err: ErrDone, args: argsStr(`prgm -a pos1 -b`), optIndex: 4},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it does not parse long opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm --longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{char: '-', err: ErrUnknownOpt, args: argsStr(`prgm --longa`), optIndex: 1},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it does not parse long opts with '-' prefix", func(t *testing.T) {
		s := NewState(argsStr(`prgm -longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{char: 'l', err: ErrUnknownOpt, args: argsStr(`prgm -longa`), optIndex: 1},
		}

		assertSeq(t, s, p, wants)
	})
}

func TestGetOpt_FuncGetOptLong(t *testing.T) {
	function := FuncGetOptLong

	t.Run("it parses short opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a p1`))
		p := Params{Opts: OptStr(`a`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a p1`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -a p1`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses long opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm --longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm --longa`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses long opts with multi-byte elements", func(t *testing.T) {
		s := NewState(argsStr(`prgm --άλφα --βήτα --γάμμα`))
		p := Params{LongOpts: LongOptStr(`άλφα,βήτα,γάμμα`), Function: function}

		wants := []assertion{
			{name: "άλφα", args: argsStr(`prgm --άλφα --βήτα --γάμμα`), optIndex: 2},
			{name: "βήτα", args: argsStr(`prgm --άλφα --βήτα --γάμμα`), optIndex: 3},
			{name: "γάμμα", args: argsStr(`prgm --άλφα --βήτα --γάμμα`), optIndex: 4},
			{err: ErrDone, args: argsStr(`prgm --άλφα --βήτα --γάμμα`), optIndex: 4},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it does not parse long opts with '-' prefix", func(t *testing.T) {
		s := NewState(argsStr(`prgm -longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{char: 'l', err: ErrUnknownOpt, args: argsStr(`prgm -longa`), optIndex: 1},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it treats arguments after '--' as parameters", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a --longa -- --longb`))
		p := Params{Opts: OptStr(`a`), LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a --longa -- --longb`), optIndex: 2},
			{name: "longa", args: argsStr(`prgm -a --longa -- --longb`), optIndex: 3},
			{err: ErrDone, args: argsStr(`prgm -a --longa -- --longb`), optIndex: 4},
		}

		assertSeq(t, s, p, wants)
	})
}

func TestGetOpt_FuncGetOptLongOnly(t *testing.T) {
	function := FuncGetOptLongOnly

	t.Run("it parses short opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a p1`))
		p := Params{Opts: OptStr(`a`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a p1`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -a p1`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses long opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm --longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm --longa`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it parses long opts with '-' prefix", func(t *testing.T) {
		s := NewState(argsStr(`prgm -longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{name: "longa", args: argsStr(`prgm -longa`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -longa`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})
}

type assertion struct {
	char     rune
	name     string
	optArg   string
	err      error
	args     []string
	optIndex int
}

func assertGetOpt(t testing.TB, s *State, p Params, want assertion) {
	t.Helper()

	res, err := s.GetOpt(p)

	if res.Char != want.char {
		t.Errorf("got Char %q, but wanted %q", res.Char, want.char)
	}
	if res.Name != want.name {
		t.Errorf("got Name %q, but wanted %q", res.Name, want.name)
	}
	if res.OptArg != want.optArg {
		t.Errorf("got OptArg %q, but wanted %q", res.OptArg, want.optArg)
	}
	if !slices.Equal(s.Args, want.args) {
		t.Errorf("got Args %v, but wanted %v", s.Args, want.args)
	}
	if s.OptIndex != want.optIndex {
		t.Errorf("got OptIndex %d, but wanted %d", s.OptIndex, want.optIndex)
	}
	if want.err == nil {
		if err != nil {
			t.Errorf("wanted no error, but got %q", err)
		}
	} else {
		if err == nil {
			t.Errorf("wanted an error, but didn't get one")
		} else if !errors.Is(err, want.err) {
			t.Errorf("got error %q, but wanted %q", err, want.err)
		}
	}
}

func assertSeq(t testing.TB, s *State, p Params, wants []assertion) {
	t.Helper()

	for _, want := range wants {
		assertGetOpt(t, s, p, want)
	}
}

func argsStr(argsStr string) (args []string) {
	// TODO: this parsing is extremely basic, maybe improve and move to module's public interface?
	var current strings.Builder
	inSingle := false
	inDouble := false

	for _, r := range argsStr {
		switch {
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case unicode.IsSpace(r) && !inSingle && !inDouble:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
