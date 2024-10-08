package getopt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"
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
		var f fixture
		if err := decoder.Decode(&f); err != nil {
			t.Fatalf("error decoding fixture: %v", err)
		}
		testName := fmt.Sprintf("%s %s %s)", f.Label, f.Func.String(), f.Mode.String())
		t.Run(testName, func(t *testing.T) {
			assertFixture(t, f)
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
		Func:     f.Func,
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

	if s.optInd != f.WantOptInd {
		t.Errorf("got optInd %d, but wanted %d", s.optInd, f.WantOptInd)
	}

	if !slices.Equal(s.args, f.WantArgs) {
		t.Errorf("got Args %+q, but wanted %+q", s.args, f.WantArgs)
	}
}

type fixtureResult struct {
	Char   rune
	Name   string
	OptArg string
	Err    error
}

type fixture struct {
	Label       string          `json:"label"`
	Func        Func            `json:"func"`
	Mode        Mode            `json:"mode"`
	Args        []string        `json:"args"`
	Opts        []Opt           `json:"opts"`
	LongOpts    []LongOpt       `json:"lopts"`
	WantArgs    []string        `json:"want_args"`
	WantOptInd  int             `json:"want_optind"`
	WantResults []fixtureResult `json:"want_results"`
}

func (f *fixture) UnmarshalJSON(data []byte) error {
	type alias fixture
	aux := &struct {
		Func        string            `json:"func"`
		Mode        string            `json:"mode"`
		Opts        []json.RawMessage `json:"opts"`
		LongOpts    []json.RawMessage `json:"lopts"`
		WantResults []json.RawMessage `json:"want_results"`
		*alias
	}{
		alias: (*alias)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch aux.Func {
	case "getopt":
		f.Func = FuncGetOpt
	case "getopt_long":
		f.Func = FuncGetOptLong
	case "getopt_long_only":
		f.Func = FuncGetOptLongOnly
	default:
		return fmt.Errorf("invalid Func: %s", aux.Func)
	}

	switch aux.Mode {
	case "gnu":
		f.Mode = ModeGNU
	case "posix":
		f.Mode = ModePosix
	case "inorder":
		f.Mode = ModeInOrder
	}

	f.Opts = make([]Opt, len(aux.Opts))
	for i, raw := range aux.Opts {
		var jsonOpt struct {
			Char   int    `json:"char"`
			HasArg string `json:"has_arg"`
		}
		err := json.Unmarshal(raw, &jsonOpt)
		if err != nil {
			return err
		}
		f.Opts[i].Char = rune(jsonOpt.Char)
		f.Opts[i].HasArg, err = parseHasArg(jsonOpt.HasArg)
		if err != nil {
			return err
		}
	}

	f.LongOpts = make([]LongOpt, len(aux.LongOpts))
	for i, raw := range aux.LongOpts {
		var jsonLongOpt struct {
			Name   string `json:"name"`
			HasArg string `json:"has_arg"`
		}
		err := json.Unmarshal(raw, &jsonLongOpt)
		if err != nil {
			return err
		}
		f.LongOpts[i].Name = jsonLongOpt.Name
		f.LongOpts[i].HasArg, err = parseHasArg(jsonLongOpt.HasArg)
		if err != nil {
			return err
		}
	}

	f.WantResults = make([]fixtureResult, len(aux.WantResults))
	for i, raw := range aux.WantResults {
		var jsonResult struct {
			Char   int    `json:"char"`
			Name   string `json:"name"`
			OptArg string `json:"optarg"`
			Err    string `json:"err"`
		}
		if err := json.Unmarshal(raw, &jsonResult); err != nil {
			return err
		}
		f.WantResults[i].Char = rune(jsonResult.Char)
		f.WantResults[i].Name = jsonResult.Name
		f.WantResults[i].OptArg = jsonResult.OptArg
		f.WantResults[i].Err = parseErr(jsonResult.Err)
	}

	return nil
}

func parseErr(err string) error {
	switch err {
	case "done":
		return ErrDone
	case "unknown_opt":
		return ErrUnknownOpt
	case "missing_opt_arg":
		return ErrMissingOptArg
	case "illegal_opt_arg":
		return ErrIllegalOptArg
	default:
		return nil
	}
}

func parseHasArg(str string) (HasArg, error) {
	switch str {
	case "no_argument":
		return NoArgument, nil
	case "required_argument":
		return RequiredArgument, nil
	case "optional_argument":
		return OptionalArgument, nil
	default:
		return 0, fmt.Errorf("invalid HasArg: %q", str)
	}
}
