package getopt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/jon-codes/getopt/internal/testgen"
)

func TestOptStr(t *testing.T) {
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
}

func TestLongOptStr(t *testing.T) {
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
}

func TestArgsStr(t *testing.T) {
	t.Run("it parses args strings", func(t *testing.T) {
		got := argsStr(`arg1 'arg2a "arg2b"' "arg3a 'arg3b'"`)
		want := []string{"arg1", "arg2a \"arg2b\"", "arg3a 'arg3b'"}

		if !slices.Equal(got, want) {
			t.Errorf("got %+q, but wanted %+q", got, want)
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

func TestGetOpt(t *testing.T) {
	function := FuncGetOpt

	t.Run("it parses short opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a p1`))
		p := Params{Opts: OptStr(`a`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a p1`), optIndex: 2},
			{err: ErrDone, args: argsStr(`prgm -a p1`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it errors on undefined opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -a`))
		p := Params{Opts: OptStr(`b`), Function: function}

		wants := []assertion{
			{char: 'a', err: ErrUnknownOpt, args: argsStr(`prgm -a`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it errors on illegal opts", func(t *testing.T) {
		s := NewState(argsStr(`prgm -π`))
		p := Params{Opts: OptStr(`π`), Function: function}

		wants := []assertion{
			{char: 'π', err: ErrUnknownOpt, args: argsStr(`prgm -π`), optIndex: 2},
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
		s := NewState(argsStr(`prgm -a -- p1`))
		p := Params{Opts: OptStr(`a::`), Function: function}

		wants := []assertion{
			{char: 'a', args: argsStr(`prgm -a -- p1`), optIndex: 3},
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
			{char: '-', err: ErrUnknownOpt, args: argsStr(`prgm --longa`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})

	t.Run("it does not parse long opts with '-' prefix", func(t *testing.T) {
		s := NewState(argsStr(`prgm -longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{char: 'l', err: ErrUnknownOpt, args: argsStr(`prgm -longa`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})
}

func TestGetOptLong(t *testing.T) {
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

	t.Run("it does not parse long opts with '-' prefix", func(t *testing.T) {
		s := NewState(argsStr(`prgm -longa`))
		p := Params{LongOpts: LongOptStr(`longa`), Function: function}

		wants := []assertion{
			{char: 'l', err: ErrUnknownOpt, args: argsStr(`prgm -longa`), optIndex: 2},
		}

		assertSeq(t, s, p, wants)
	})
}

func TestGetOptLongOnly(t *testing.T) {
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

const fixturePath = "testdata/fixtures.json"

func TestGetOpt_Fixtures(t *testing.T) {
	fixtureFile, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixtures file: %v", err)
	}
	defer fixtureFile.Close()

	decoder := json.NewDecoder(fixtureFile)

	// read open bracket
	_, err = decoder.Token()
	if err != nil {
		t.Fatalf("error decoding cases: %v", err)
	}

	// while the array contains values
	for decoder.More() {
		var record testgen.FixtureRecord
		if err := decoder.Decode(&record); err != nil {
			t.Fatalf("error decoding fixture: %v", err)
		}
		fixture, err := buildFixture(record)
		if err != nil {
			t.Fatalf("error parsing fixture: %v", err)
		}
		testName := fmt.Sprintf("Fixture %q (function %q, mode %q)", record.Label, record.FunctionStr, record.ModeStr)
		t.Run(testName, func(t *testing.T) {
			assertFixture(t, fixture)
		})
	}

	// read closing bracket
	_, err = decoder.Token()
	if err != nil {
		t.Fatalf("error decoding cases: %v", err)
	}
}

func assertFixture(t testing.TB, f fixture) {
	t.Helper()

	s := NewState(f.Args)
	p := Params{
		Opts:     f.Opts,
		LongOpts: f.LongOpts,
		Mode:     f.Mode,
		Function: f.Function,
	}

	for iter, want := range f.WantResults {
		res, err := s.GetOpt(p)

		if want.Err == nil {
			if err != nil {
				t.Errorf("iter %d, wanted no error, but got %q", iter, err)
			}
		} else {
			if err == nil {
				t.Errorf("iter %d, wanted an error, but didn't get one", iter)
			} else if !errors.Is(err, want.Err) {
				t.Errorf("iter %d, got error %q, but wanted %q", iter, err, want.Err)
			}
		}

		if res.Char != want.Char {
			t.Errorf("iter %d, wanted Char %q, but got %q", iter, res.Char, want.Char)
		}
		if res.Name != want.Name {
			t.Errorf("iter %d, wanted Name %q, but got %q", iter, res.Name, want.Name)
		}
		if res.OptArg != want.OptArg {
			t.Errorf("iter %d, wanted OptArg %q, but got %q", iter, res.OptArg, want.OptArg)
		}
	}

	if s.OptIndex != f.WantOptIndex {
		t.Errorf("got OptIndex %d, but wanted %d", s.OptIndex, f.WantOptIndex)
	}

	if !slices.Equal(s.Args, f.WantArgs) {
		t.Errorf("got Args %+q, but wanted %+q", s.Args, f.WantArgs)
	}
}

type fixtureIter struct {
	Char   rune
	Name   string
	OptArg string
	Err    error
}

type fixture struct {
	Label        string
	Args         []string
	Opts         []Opt
	LongOpts     []LongOpt
	Function     GetOptFunc
	Mode         GetOptMode
	WantArgs     []string
	WantOptIndex int
	WantResults  []fixtureIter
}

func buildFixture(fr testgen.FixtureRecord) (f fixture, err error) {
	var function GetOptFunc
	switch fr.FunctionStr {
	case "getopt":
		function = FuncGetOpt
	case "getopt_long":
		function = FuncGetOptLong
	case "getopt_long_only":
		function = FuncGetOptLongOnly
	default:
		return f, fmt.Errorf("unknown function type %q", fr.FunctionStr)
	}

	var mode GetOptMode
	switch fr.ModeStr {
	case "gnu":
		mode = ModeGNU
	case "posix":
		mode = ModePosix
	case "inorder":
		mode = ModeInOrder
	default:
		return f, fmt.Errorf("unknown mode type %q", fr.ModeStr)
	}

	var wantResults []fixtureIter
	for _, fi := range fr.WantResults {
		var char rune
		if fi.CharStr != "" {
			char, _ = utf8.DecodeRuneInString(fi.CharStr)
		}

		var err error
		switch fi.ErrStr {
		case "":
			err = nil
		case "-1":
			err = ErrDone
		case ":":
			err = ErrMissingOptArg
		case "?":
			err = ErrUnknownOpt
		default:
			return f, fmt.Errorf("unknown error type %q", fi.ErrStr)
		}

		wantResults = append(wantResults, fixtureIter{
			Char:   char,
			Name:   fi.Name,
			OptArg: fi.OptArg,
			Err:    err,
		})
	}

	f.Label = fr.Label
	f.Args = argsStr(fr.ArgsStr)
	f.Opts = OptStr(fr.OptStr)
	f.LongOpts = LongOptStr(fr.LongOptStr)
	f.Function = function
	f.Mode = mode
	f.WantArgs = argsStr(fr.WantArgsStr)
	f.WantOptIndex = fr.WantOptIndex
	f.WantResults = wantResults

	return f, nil
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
