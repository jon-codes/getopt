package opt

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	g := NewGetOptState([]string{"my_program", "-a", "-b"})
	assertState(t, g, GetOptState{Args: []string{"my_program", "-a", "-b"}, OptIndex: 1, ArgIndex: 0})
}

func TestReset(t *testing.T) {
	g := NewGetOptState([]string{"my_program", "-a", "-b"})
	p := GetOptParams{Opts: []Opt{{Char: 'a'}, {Char: 'b'}}}

	_, _ = g.GetOpt(p)
	assertState(t, g, GetOptState{Args: []string{"my_program", "-a", "-b"}, OptIndex: 2, ArgIndex: 0})

	g.Reset([]string{"my_program", "-b"})
	assertState(t, g, GetOptState{Args: []string{"my_program", "-b"}, OptIndex: 1, ArgIndex: 0})
}

func TestGetOpt_Opts(t *testing.T) {
	t.Run("a single valid option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a"})
		p := GetOptParams{Opts: []Opt{{Char: 'a'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a'}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a"}, 2)
	})

	t.Run("a single valid option when the argument is provided inline", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-afoo"})
		p := GetOptParams{Opts: []Opt{{Char: 'a', HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a', OptArg: "foo"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-afoo"}, 2)
	})

	t.Run("a single valid option when the argument is provided in the next arg", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "foo"})
		p := GetOptParams{Opts: []Opt{{Char: 'a', HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a', OptArg: "foo"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "foo"}, 3)
	})

	t.Run("a single valid option when the next arg looks like an option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "-b"})
		p := GetOptParams{Opts: []Opt{{Char: 'a', HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a', OptArg: "-b"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "-b"}, 3)
	})

	t.Run("a single valid option when the argument contains multi-byte chars", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a文"})
		p := GetOptParams{Opts: []Opt{{Char: 'a', HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a', OptArg: "文"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a文"}, 2)
	})

	t.Run("multiple valid options", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "-b", "-c"})
		p := GetOptParams{Opts: []Opt{{Char: 'a'}, {Char: 'b'}, {Char: 'c'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a'}, err: nil},
			{res: GetOptResult{Char: 'b'}, err: nil},
			{res: GetOptResult{Char: 'c'}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "-b", "-c"}, 4)
	})

	t.Run("option groups", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-abc"})
		p := GetOptParams{Opts: []Opt{{Char: 'a'}, {Char: 'b'}, {Char: 'c'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a'}, err: nil},
			{res: GetOptResult{Char: 'b'}, err: nil},
			{res: GetOptResult{Char: 'c'}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-abc"}, 2)
	})

	t.Run("a single undefined option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-b"})
		p := GetOptParams{Opts: []Opt{{Char: 'a'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'b'}, err: ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-b"}, 1)
	})

	t.Run("a single '-' option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-"})
		p := GetOptParams{Opts: []Opt{{Char: '-'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: '-'}, err: ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-"}, 1)
	})

	t.Run("a single non-graph option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-文"})
		p := GetOptParams{Opts: []Opt{{Char: '文'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: '文'}, err: ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-文"}, 1)
	})

	t.Run("-- terminates option parsing", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "--", "-b"})
		p := GetOptParams{Opts: []Opt{{Char: 'a'}, {Char: 'b'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a'}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "--", "-b"}, 3)
	})

	t.Run("positionals terminate option parsing", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "my_arg", "-b"})
		p := GetOptParams{Opts: []Opt{{Char: 'a'}, {Char: 'b'}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a'}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "my_arg", "-b"}, 2)
	})

	t.Run("a single option with missing required argument", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a"})
		p := GetOptParams{Opts: []Opt{{Char: 'a', HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a'}, err: ErrMissingOptArg},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a"}, 2)
	})

	t.Run("a single option with missing optional argument", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a"})
		p := GetOptParams{Opts: []Opt{{Char: 'a', HasArg: OptionalArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Char: 'a'}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a"}, 2)
	})
}

func TestGetOpt_LongOpts(t *testing.T) {
	t.Run("a single valid option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo"}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo"}, 2)
	})

	t.Run("a single valid option when the argument is provided inline", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo=bar"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo", HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo", OptArg: "bar"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo=bar"}, 2)
	})

	t.Run("a single valid option when the argument is provided in the next arg", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo", "bar"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo", HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo", OptArg: "bar"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo", "bar"}, 3)
	})

	t.Run("a single valid option when the next arg looks like an option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo", "--bar"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo", HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo", OptArg: "--bar"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo", "--bar"}, 3)
	})

	t.Run("a single valid option when the argument contains multi-byte chars", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo=文"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo", HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo", OptArg: "文"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo=文"}, 2)
	})

	t.Run("a single option containing invalid chars", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo文"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo文"}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo文"}, err: ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo文"}, 2)
	})

	t.Run("a single undefined option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "bar"}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo"}, err: ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo"}, 2)
	})

	t.Run("a single option with a disallowed inline option argument", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo=bar"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo"}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo=bar"}, err: ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo=bar"}, 2)
	})

	t.Run("a single option with missing required argument", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo", HasArg: RequiredArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo"}, err: ErrMissingOptArg},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo"}, 2)
	})

	t.Run("a single option with missing optional argument", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "--foo"})
		p := GetOptParams{LongOpts: []LongOpt{{Name: "foo", HasArg: OptionalArgument}}}

		wants := []resultAssertion{
			{res: GetOptResult{Name: "foo"}, err: nil},
			{res: GetOptResult{}, err: ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "--foo"}, 2)
	})
}

func TestGetOpt_MixedOpts(t *testing.T) {
	g := NewGetOptState([]string{"my_program", "-abc", "d", "--foo=bar", "-ef", "--", "--fizz", "buzz"})
	p := GetOptParams{
		Opts: []Opt{
			{Char: 'a'},
			{Char: 'b'},
			{Char: 'c', HasArg: RequiredArgument},
			{Char: 'e', HasArg: OptionalArgument},
		}, LongOpts: []LongOpt{
			{Name: "foo", HasArg: RequiredArgument},
		}}

	wants := []resultAssertion{
		{res: GetOptResult{Char: 'a'}, err: nil},
		{res: GetOptResult{Char: 'b'}, err: nil},
		{res: GetOptResult{Char: 'c', OptArg: "d"}, err: nil},
		{res: GetOptResult{Name: "foo", OptArg: "bar"}, err: nil},
		{res: GetOptResult{Char: 'e', OptArg: "f"}, err: nil},
		{res: GetOptResult{}, err: ErrDone},
	}

	assertSequence(t, g, p, wants)
	assertArgs(t, g, []string{"my_program", "-abc", "d", "--foo=bar", "-ef", "--", "--fizz", "buzz"}, 6)
}

func TestGetOpt_Fixtures(t *testing.T) {
	f, err := os.Open("testdata/fixtures.csv")
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)

	headers, err := reader.Read()
	if err != nil {
		t.Fatalf("error reading file headers: %v", err)
	}
	colMap, err := readFixtureHeaders(headers)
	if err != nil {
		t.Fatalf("error reading file headers: %v", err)
	}

	var g *GetOptState

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("error reading record: %v", err)
		}
		f, err := readFixtureRow(row, colMap)
		if err != nil {
			t.Fatalf("error parsing fixture: %v", err)
		}

		if f.step == 0 {
			g = NewGetOptState(f.args)
		}

		testName := fmt.Sprintf("%s, %s, %s, iter %d", f.label, f.function.String(), f.mode.String(), f.step)
		t.Run(testName, func(t *testing.T) {
			got, err := f.got(g)
			want := f.want()

			if f.err == nil {
				assertNoError(t, err)
			} else {
				assertError(t, err, f.err)
			}
			assertResult(t, got, want)
			assertState(t, g, GetOptState{Args: f.mutargs, OptIndex: f.optindex})
		})
	}
}

type resultAssertion struct {
	res GetOptResult
	err error
}

func assertSequence(t testing.TB, g *GetOptState, p GetOptParams, wants []resultAssertion) {
	t.Helper()

	for _, want := range wants {
		got, err := g.GetOpt(p)
		if want.err == nil {
			assertNoError(t, err)
		} else {
			assertError(t, err, want.err)
		}
		assertResult(t, got, want.res)
	}
}

func assertResult(t testing.TB, got GetOptResult, want GetOptResult) {
	t.Helper()

	if got.Char != want.Char {
		t.Errorf("got Char %q, but wanted %q", got.Char, want.Char)
	}
	if got.Name != want.Name {
		t.Errorf("got Name %q, but wanted %q", got.Name, want.Name)
	}
	if got.OptArg != want.OptArg {
		t.Errorf("got OptArg %q, but wanted %q", got.OptArg, want.OptArg)
	}
}

func assertArgs(t testing.TB, g *GetOptState, wantArgs []string, wantOptIndex int) {
	t.Helper()

	if !slices.Equal(g.Args, wantArgs) {
		t.Errorf("got Args %v, but wanted %v", g.Args, wantArgs)
	}
	if g.OptIndex != wantOptIndex {
		t.Errorf("got OptIndex %d, but wanted %d", g.OptIndex, wantOptIndex)
	}
}

func assertState(t testing.TB, got *GetOptState, want GetOptState) {
	t.Helper()

	if !slices.Equal(got.Args, want.Args) {
		t.Errorf("got Args %v, but wanted %v", got.Args, want.Args)
	}
	if got.OptIndex != want.OptIndex {
		t.Errorf("got OptIndex %d, but wanted %d", got.OptIndex, want.OptIndex)
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()

	if got == nil {
		t.Fatal("wanted an error, but didn't get one")
	}

	if !errors.Is(got, want) {
		t.Errorf("got error %q, but wanted %q", got, want)
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("wanted no error, but got %q", err)
	}
}

type fixture struct {
	label    string
	step     int
	args     []string
	opts     []Opt
	longOpts []LongOpt
	function GetOptFunc
	mode     GetOptMode
	char     rune
	name     string
	err      error
	optindex int
	optarg   string
	mutargs  []string
}

func (f fixture) got(g *GetOptState) (GetOptResult, error) {
	p := GetOptParams{Opts: f.opts, LongOpts: f.longOpts, GetOptFunc: f.function, Mode: f.mode}
	return g.GetOpt(p)
}

func (f fixture) want() GetOptResult {
	return GetOptResult{Char: f.char, Name: f.name, OptArg: f.optarg}
}

var colNames = []string{
	"label", "step", "args", "optstring", "longopts", "char", "name", "err",
	"optindex", "optarg", "mutargs",
}

func readFixtureHeaders(row []string) (map[string]int, error) {
	m := map[string]int{}
	for _, name := range colNames {
		i := slices.Index(row, name)
		if i < 0 {
			return m, fmt.Errorf("could not find column name %s in row %v", name, row)
		}
		m[name] = i
	}
	return m, nil
}

func readFixtureRow(row []string, cols map[string]int) (fixture, error) {
	labelStr := row[cols["label"]]
	stepStr := row[cols["step"]]
	argsStr := row[cols["args"]]
	optstringStr := row[cols["optstring"]]
	longoptsStr := row[cols["longopts"]]
	charStr := row[cols["char"]]
	nameStr := row[cols["name"]]
	errStr := row[cols["err"]]
	optindexStr := row[cols["optindex"]]
	optargStr := row[cols["optarg"]]
	mutargsStr := row[cols["mutargs"]]

	step, err := strconv.Atoi(stepStr)
	if err != nil {
		return fixture{}, fmt.Errorf("error parsing fixture step: %v", err)
	}

	var args []string
	err = json.Unmarshal([]byte(argsStr), &args)
	if err != nil {
		return fixture{}, fmt.Errorf("error parsing fixture JSON args: %v", err)
	}

	var opts []Opt

	for i := 0; i < len(optstringStr); i++ {
		char := rune(optstringStr[i])
		hasArg := NoArgument

		if i+1 < len(optstringStr) && optstringStr[i+1] == ':' {
			hasArg = RequiredArgument
			i++
			if i+1 < len(optstringStr) && optstringStr[i+1] == ':' {
				hasArg = OptionalArgument
				i++
			}
		}
		opts = append(opts, Opt{Char: char, HasArg: hasArg})
	}

	var longOpts []LongOpt
	items := strings.Split(longoptsStr, ",")

	for _, item := range items {
		hasArg := NoArgument
		name, found := strings.CutSuffix(item, "::")
		if found {
			hasArg = OptionalArgument
		} else {
			name, found = strings.CutSuffix(item, ":")
			if found {
				hasArg = RequiredArgument
			}
		}

		longOpts = append(longOpts, LongOpt{Name: name, HasArg: hasArg})
	}

	var mutargs []string
	err = json.Unmarshal([]byte(mutargsStr), &mutargs)
	if err != nil {
		return fixture{}, fmt.Errorf("error parsing fixture JSON mutable args: %v", err)
	}

	char := '\x00'
	if charStr != "" {
		char = rune(charStr[0])
	}

	var errVal error
	if errStr == "?" {
		errVal = ErrIllegalOpt
	}
	if errStr == ":" {
		errVal = ErrMissingOptArg
	}
	if errStr == "-1" {
		errVal = ErrDone
	}

	optindex, err := strconv.Atoi(optindexStr)
	if err != nil {
		return fixture{}, fmt.Errorf("error parsing fixture optindex: %v", err)
	}

	return fixture{
		label:    labelStr,
		step:     step,
		args:     args,
		opts:     opts,
		longOpts: longOpts,
		char:     char,
		name:     nameStr,
		err:      errVal,
		optindex: optindex,
		optarg:   optargStr,
		mutargs:  mutargs,
	}, nil
}
