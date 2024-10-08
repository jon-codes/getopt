package getopt

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"unicode"
)

func TestParse(t *testing.T) {
	t.Run("with empty results", func(t *testing.T) {
		s := testState(`prgm p1 p2 p3`)
		c := Config{Opts: OptStr(`abc`)}
		got, err := s.Parse(c)
		if err != nil {
			t.Fatalf("got error, but didn't expect one")
		}
		if len(got) != 0 {
			t.Errorf("got len %d, but wanted %d", len(got), 0)
		}

		wantParams := argsStr(`p1 p2 p3`)

		if !slices.Equal(s.Params(), wantParams) {
			t.Errorf("got %+v, but wanted %+v", s.Params(), wantParams)
		}
	})

	t.Run("with results", func(t *testing.T) {
		s := testState(`prgm -a -bc p1 p2 p3`)
		c := Config{Opts: OptStr(`abc`)}
		got, err := s.Parse(c)
		want := []Result{
			{Char: 'a'},
			{Char: 'b'},
			{Char: 'c'},
		}

		if err != nil {
			t.Fatalf("got error, but didn't expect one")
		}
		if !slices.Equal(got, want) {
			t.Errorf("got %+v, but wanted %+v", got, want)
		}

		wantParams := argsStr(`p1 p2 p3`)

		if !slices.Equal(s.Params(), wantParams) {
			t.Errorf("got %+v, but wanted %+v", s.Params(), wantParams)
		}
	})

	t.Run("with error", func(t *testing.T) {
		s := testState(`prgm -a -d -bc p1 p2 p3`)
		c := Config{Opts: OptStr(`abc`)}
		got, err := s.Parse(c)
		want := []Result{
			{Char: 'a'},
		}

		if err != ErrUnknownOpt {
			t.Fatalf("got no error, but expected %v", ErrUnknownOpt)
		}
		if !slices.Equal(got, want) {
			t.Errorf("got %+v, but wanted %+v", got, want)
		}

		wantParams := argsStr(`-bc p1 p2 p3`)

		if !slices.Equal(s.Params(), wantParams) {
			t.Errorf("got %+v, but wanted %+v", s.Params(), wantParams)
		}
	})
}

func TestAll(t *testing.T) {
	s := testState(`prgm -a -dc`)
	c := Config{Opts: OptStr(`abc`)}
	want := []struct {
		res Result
		err error
	}{
		{res: Result{Char: 'a'}, err: nil},
		{res: Result{Char: 'd'}, err: ErrUnknownOpt},
		{res: Result{Char: 'c'}, err: nil},
		{res: Result{}, err: ErrDone},
	}

	i := 0
	for opt, err := range s.All(c) {
		if want[i].err != nil {
			if err == nil {
				t.Fatalf("got no error, but wanted one")
			}
			if err != want[i].err {
				t.Fatalf("got error %v, but wanted %v", err, want[i].err)
			}
		} else {
			if err != nil {
				t.Fatalf("got err %v, but wanted none", err)
			}
		}
		if opt.Char != want[i].res.Char {
			t.Fatalf("got Char %q, but wanted %q", opt.Char, want[i].res.Char)
		}
		i++
	}
}

func TestNew(t *testing.T) {
	got := NewState(argsStr(`prgm -a -b`))
	want := State{optInd: 1, args: []string{"prgm", "-a", "-b"}}

	if got.optInd != want.optInd {
		t.Errorf("got %d, but wanted %d", got.optInd, want.optInd)
	}

	if !slices.Equal(got.args, want.args) {
		t.Errorf("got %+q, but wanted %+q", got.args, want.args)
	}
}

func TestArgs(t *testing.T) {
	s := testState(`prgm -a -bc`)
	got := s.Args()

	if !slices.Equal(got, s.args) {
		t.Errorf("got %+q, but wanted %+q", got, s.Args())
	}
}

func TestOptInd(t *testing.T) {
	s := NewState(argsStr(`prgm -a -bc`))
	got := s.OptInd()

	if got != s.optInd {
		t.Errorf("got %d, but wanted %d", got, s.optInd)
	}
}

func TestParams(t *testing.T) {
	t.Run("when params are present", func(t *testing.T) {
		s := testState(`prgm -a p1 -bc p2 p3`)
		c := Config{Opts: OptStr(`abc`)}

		for range s.All(c) {
		}
		got := s.Params()
		want := []string{"p1", "p2", "p3"}

		if !slices.Equal(got, want) {
			t.Errorf("got %+q, but wanted %+q", got, want)
		}
	})

	t.Run("when params are not present", func(t *testing.T) {
		s := testState(`prgm -a -bc`)
		c := Config{Opts: OptStr(`abc`)}

		for range s.All(c) {
		}
		got := len(s.Params())

		if got != 0 {
			t.Errorf("got len %d, but wanted %d", got, 0)
		}
	})
}

func TestReset(t *testing.T) {
	s := NewState(argsStr(`prgm -ab -c`))
	c := Config{Opts: OptStr(`abc`)}

	assertGetOpt(t, s, c, assertion{
		char:   'a',
		args:   argsStr(`prgm -ab -c`),
		optInd: 1,
	})

	s.Reset(argsStr(`prgm -c -b`))

	assertGetOpt(t, s, c, assertion{
		char:   'c',
		args:   argsStr(`prgm -c -b`),
		optInd: 2,
	})
}

func testState(args string) *State {
	return NewState(argsStr(args))
}

func TestGetOpt_FuncGetOpt(t *testing.T) {
	function := FuncGetOpt

	t.Run("it parses short opts", func(t *testing.T) {
		s := testState(`prgm -a -bc`)
		c := Config{
			Opts: OptStr(`abc`),
			Func: function,
			Mode: ModeGNU,
		}
		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'b', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'c', args: argsStr(`prgm -a -bc`), optInd: 3},
			{err: ErrDone, args: argsStr(`prgm -a -bc`), optInd: 3},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses short opts with required args", func(t *testing.T) {
		s := testState(`prgm -aarg1 -b arg2 -c`)
		c := Config{
			Opts: OptStr(`a:b:c:`),
			Func: function,
			Mode: ModeGNU,
		}
		wants := []assertion{
			{char: 'a', optArg: "arg1", args: argsStr(`prgm -aarg1 -b arg2 -c`), optInd: 2},
			{char: 'b', optArg: "arg2", args: argsStr(`prgm -aarg1 -b arg2 -c`), optInd: 4},
			{char: 'c', err: ErrMissingOptArg, args: argsStr(`prgm -aarg1 -b arg2 -c`), optInd: 5},
			{err: ErrDone, args: argsStr(`prgm -aarg1 -b arg2 -c`), optInd: 5},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses short opts with optional args", func(t *testing.T) {
		s := testState(`prgm -aarg1 -b -c`)
		c := Config{
			Opts: OptStr(`a::b::c::`),
			Func: function,
			Mode: ModeGNU,
		}
		wants := []assertion{
			{char: 'a', optArg: "arg1", args: argsStr(`prgm -aarg1 -b -c`), optInd: 2},
			{char: 'b', args: argsStr(`prgm -aarg1 -b -c`), optInd: 3},
			{char: 'c', args: argsStr(`prgm -aarg1 -b -c`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm -aarg1 -b -c`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it permutes parameters in gnu mode", func(t *testing.T) {
		s := testState(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`)
		c := Config{
			Opts: OptStr(`ab:c::`),
			Func: function,
			Mode: ModeGNU,
		}
		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 2},
			{char: 'b', optArg: "arg1", args: argsStr(`prgm -a -b arg1 p1 p2 p3 p4 -c -- p5`), optInd: 4},
			{char: 'c', args: argsStr(`prgm -a -b arg1 -c p1 p2 p3 p4 -- p5`), optInd: 5},
			{err: ErrDone, args: argsStr(`prgm -a -b arg1 -c -- p1 p2 p3 p4 p5`), optInd: 6},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it does not permute parameters in posix mode", func(t *testing.T) {
		s := testState(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`)
		c := Config{
			Opts: OptStr(`ab:c::`),
			Func: function,
			Mode: ModePosix,
		}
		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 2},
			{err: ErrDone, args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 2},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it treats parameters as options in inorder mode", func(t *testing.T) {
		s := testState(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`)
		c := Config{
			Opts: OptStr(`ab:c::`),
			Func: function,
			Mode: ModeInOrder,
		}
		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 2},
			{char: 1, optArg: "p1", args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 3},
			{char: 1, optArg: "p2", args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 4},
			{char: 'b', optArg: "arg1", args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 6},
			{char: 1, optArg: "p3", args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 7},
			{char: 1, optArg: "p4", args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 8},
			{char: 'c', args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 9},
			{err: ErrDone, args: argsStr(`prgm -a p1 p2 -b arg1 p3 p4 -c -- p5`), optInd: 10},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it repeats when done", func(t *testing.T) {
		s := testState(`prgm -a -bc`)
		c := Config{
			Opts: OptStr(`abc`),
			Func: function,
			Mode: ModeGNU,
		}
		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'b', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'c', args: argsStr(`prgm -a -bc`), optInd: 3},
			{err: ErrDone, args: argsStr(`prgm -a -bc`), optInd: 3},
			{err: ErrDone, args: argsStr(`prgm -a -bc`), optInd: 3},
			{err: ErrDone, args: argsStr(`prgm -a -bc`), optInd: 3},
		}

		assertSeq(t, s, c, wants)
	})
}

func TestGetOpt_FuncGetOptLong(t *testing.T) {
	function := FuncGetOptLong

	t.Run("it parses short opts", func(t *testing.T) {
		s := testState(`prgm -a -bc`)
		c := Config{
			Opts: OptStr(`abc`),
			Func: function,
			Mode: ModeGNU,
		}
		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'b', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'c', args: argsStr(`prgm -a -bc`), optInd: 3},
			{err: ErrDone, args: argsStr(`prgm -a -bc`), optInd: 3},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses long opts", func(t *testing.T) {
		s := testState(`prgm --longa --longb --longc`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb,longc`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa --longb --longc`), optInd: 2},
			{name: "longb", args: argsStr(`prgm --longa --longb --longc`), optInd: 3},
			{name: "longc", args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses long opts with required arguments", func(t *testing.T) {
		s := testState(`prgm --longa arg1 --longb=arg2 --longc`)
		c := Config{
			LongOpts: LongOptStr(`longa:,longb:,longc:`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", optArg: "arg1", args: argsStr(`prgm --longa arg1 --longb=arg2 --longc`), optInd: 3},
			{name: "longb", optArg: "arg2", args: argsStr(`prgm --longa arg1 --longb=arg2 --longc`), optInd: 4},
			{name: "longc", err: ErrMissingOptArg, args: argsStr(`prgm --longa arg1 --longb=arg2 --longc`), optInd: 5},
			{err: ErrDone, args: argsStr(`prgm --longa arg1 --longb=arg2 --longc`), optInd: 5},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses long opts with optional arguments", func(t *testing.T) {
		s := testState(`prgm --longa p1 --longb=arg1 --longc`)
		c := Config{
			LongOpts: LongOptStr(`longa::,longb::,longc::`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa p1 --longb=arg1 --longc`), optInd: 2},
			{name: "longb", optArg: "arg1", args: argsStr(`prgm --longa --longb=arg1 p1 --longc`), optInd: 3},
			{name: "longc", args: argsStr(`prgm --longa --longb=arg1 --longc p1`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb=arg1 --longc p1`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it permutes parameters in gnu mode", func(t *testing.T) {
		s := testState(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb:,longc::`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 2},
			{name: "longb", optArg: "arg1", args: argsStr(`prgm --longa --longb arg1 p1 p2 p3 p4 --longc -- p5`), optInd: 4},
			{name: "longc", args: argsStr(`prgm --longa --longb arg1 --longc p1 p2 p3 p4 -- p5`), optInd: 5},
			{err: ErrDone, args: argsStr(`prgm --longa --longb arg1 --longc -- p1 p2 p3 p4 p5`), optInd: 6},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it does not permute parameters in posix mode", func(t *testing.T) {
		s := testState(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb:,longc::`),
			Func:     function,
			Mode:     ModePosix,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 2},
			{err: ErrDone, args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 2},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it treats parameters as options in inorder mode", func(t *testing.T) {
		s := testState(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb:,longc::`),
			Func:     function,
			Mode:     ModeInOrder,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 2},
			{char: 1, optArg: "p1", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 3},
			{char: 1, optArg: "p2", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 4},
			{name: "longb", optArg: "arg1", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 6},
			{char: 1, optArg: "p3", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 7},
			{char: 1, optArg: "p4", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 8},
			{name: "longc", args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 9},
			{err: ErrDone, args: argsStr(`prgm --longa p1 p2 --longb arg1 p3 p4 --longc -- p5`), optInd: 10},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it repeats when done", func(t *testing.T) {
		s := testState(`prgm --longa --longb --longc`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb,longc`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa --longb --longc`), optInd: 2},
			{name: "longb", args: argsStr(`prgm --longa --longb --longc`), optInd: 3},
			{name: "longc", args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})
}

func TestGetOpt_FuncGetOptLongOnly(t *testing.T) {
	function := FuncGetOptLongOnly

	t.Run("it parses short opts", func(t *testing.T) {
		s := testState(`prgm -a -bc`)
		c := Config{
			Opts: OptStr(`abc`),
			Func: function,
			Mode: ModeGNU,
		}
		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'b', args: argsStr(`prgm -a -bc`), optInd: 2},
			{char: 'c', args: argsStr(`prgm -a -bc`), optInd: 3},
			{err: ErrDone, args: argsStr(`prgm -a -bc`), optInd: 3},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses long opts", func(t *testing.T) {
		s := testState(`prgm --longa --longb --longc`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb,longc`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa --longb --longc`), optInd: 2},
			{name: "longb", args: argsStr(`prgm --longa --longb --longc`), optInd: 3},
			{name: "longc", args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses shortened long opts", func(t *testing.T) {
		s := testState(`prgm -longa -longb -longc`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb,longc`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm -longa -longb -longc`), optInd: 2},
			{name: "longb", args: argsStr(`prgm -longa -longb -longc`), optInd: 3},
			{name: "longc", args: argsStr(`prgm -longa -longb -longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm -longa -longb -longc`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses shortened long opts with required arguments", func(t *testing.T) {
		s := testState(`prgm -longa arg1 -longb=arg2 -longc`)
		c := Config{
			LongOpts: LongOptStr(`longa:,longb:,longc:`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", optArg: "arg1", args: argsStr(`prgm -longa arg1 -longb=arg2 -longc`), optInd: 3},
			{name: "longb", optArg: "arg2", args: argsStr(`prgm -longa arg1 -longb=arg2 -longc`), optInd: 4},
			{name: "longc", err: ErrMissingOptArg, args: argsStr(`prgm -longa arg1 -longb=arg2 -longc`), optInd: 5},
			{err: ErrDone, args: argsStr(`prgm -longa arg1 -longb=arg2 -longc`), optInd: 5},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it parses long opts with optional arguments", func(t *testing.T) {
		s := testState(`prgm -longa p1 -longb=arg1 -longc`)
		c := Config{
			LongOpts: LongOptStr(`longa::,longb::,longc::`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm -longa p1 -longb=arg1 -longc`), optInd: 2},
			{name: "longb", optArg: "arg1", args: argsStr(`prgm -longa -longb=arg1 p1 -longc`), optInd: 3},
			{name: "longc", args: argsStr(`prgm -longa -longb=arg1 -longc p1`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm -longa -longb=arg1 -longc p1`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it permutes parameters in gnu mode", func(t *testing.T) {
		s := testState(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb:,longc::`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 2},
			{name: "longb", optArg: "arg1", args: argsStr(`prgm -longa -longb arg1 p1 p2 p3 p4 -longc -- p5`), optInd: 4},
			{name: "longc", args: argsStr(`prgm -longa -longb arg1 -longc p1 p2 p3 p4 -- p5`), optInd: 5},
			{err: ErrDone, args: argsStr(`prgm -longa -longb arg1 -longc -- p1 p2 p3 p4 p5`), optInd: 6},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it does not permute parameters in posix mode", func(t *testing.T) {
		s := testState(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb:,longc::`),
			Func:     function,
			Mode:     ModePosix,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 2},
			{err: ErrDone, args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 2},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it treats parameters as options in inorder mode", func(t *testing.T) {
		s := testState(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb:,longc::`),
			Func:     function,
			Mode:     ModeInOrder,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 2},
			{char: 1, optArg: "p1", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 3},
			{char: 1, optArg: "p2", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 4},
			{name: "longb", optArg: "arg1", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 6},
			{char: 1, optArg: "p3", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 7},
			{char: 1, optArg: "p4", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 8},
			{name: "longc", args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 9},
			{err: ErrDone, args: argsStr(`prgm -longa p1 p2 -longb arg1 p3 p4 -longc -- p5`), optInd: 10},
		}

		assertSeq(t, s, c, wants)
	})

	t.Run("it repeats when done", func(t *testing.T) {
		s := testState(`prgm --longa --longb --longc`)
		c := Config{
			LongOpts: LongOptStr(`longa,longb,longc`),
			Func:     function,
			Mode:     ModeGNU,
		}
		wants := []assertion{
			{name: "longa", args: argsStr(`prgm --longa --longb --longc`), optInd: 2},
			{name: "longb", args: argsStr(`prgm --longa --longb --longc`), optInd: 3},
			{name: "longc", args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
			{err: ErrDone, args: argsStr(`prgm --longa --longb --longc`), optInd: 4},
		}

		assertSeq(t, s, c, wants)
	})
}

type assertion struct {
	char   rune
	name   string
	optArg string
	err    error
	args   []string
	optInd int
}

func assertGetOpt(t testing.TB, s *State, p Config, want assertion) {
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
	if !slices.Equal(s.args, want.args) {
		t.Errorf("got Args %v, but wanted %v", s.args, want.args)
	}
	if s.optInd != want.optInd {
		t.Errorf("got optInd %d, but wanted %d", s.optInd, want.optInd)
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

func assertSeq(t testing.TB, s *State, p Config, wants []assertion) {
	t.Helper()

	for _, want := range wants {
		assertGetOpt(t, s, p, want)
	}
}

func argsStr(argsStr string) (args []string) {
	// TODO: this shlex is extremely basic and is only meant as a test helper
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
