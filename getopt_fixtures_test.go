package getopt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"
	"unicode/utf8"

	"github.com/jon-codes/getopt/internal/testgen"
)

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
			t.Errorf("iter %d, got Char %q, but wanted %q", iter, res.Char, want.Char)
		}
		if res.Name != want.Name {
			t.Errorf("iter %d, got Name %q, but wanted %q", iter, res.Name, want.Name)
		}
		if res.OptArg != want.OptArg {
			t.Errorf("iter %d, got OptArg %q, but wanted %q", iter, res.OptArg, want.OptArg)
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
			_, found := findLongOpt(fi.Name, false, Params{LongOpts: LongOptStr(fr.LongOptStr), Function: function, Mode: mode})
			if fi.Name != "" && found {
				err = ErrIllegalOptArg
			} else {
				err = ErrUnknownOpt
			}
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
