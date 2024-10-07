package getopt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

const fixturePath = "testdata/fixtures.json"

type fixtureRecordIter struct {
	Opt       int    `json:"opt"`
	OptInd    int    `json:"optind"`
	OptOpt    int    `json:"optopt"`
	OptArg    string `json:"optarg"`
	LongIndex int    `json:"longindex"`
}

type fixtureIter struct {
	Char   rune
	Name   string
	OptArg string
	Err    error
}

func (fri *fixtureRecordIter) toFixtureIter(fixture *fixture) fixtureIter {
	fi := fixtureIter{
		OptArg: fri.OptArg,
	}

	switch fri.Opt {
	case ':':
		fi.Err = ErrUnknownOpt
	case '?':
		fi.Err = ErrUnknownOpt
		if fri.OptOpt > 0 {
			fi.Char = rune(fri.OptOpt)
		} else if fixture.Function != FuncGetOpt {
			name := strings.TrimLeft(fixture.Args[fri.OptInd-1], "-")
			name = strings.SplitN(name, "=", 2)[0]
			fi.Name = name
		}
	case -1: // done
		fi.Err = ErrDone
	case -2:
		fi.Name = fixture.LongOpts[fri.LongIndex].Name
	default:
		fi.Char = rune(fri.Opt)
	}

	return fi
}

type fixtureRecord struct {
	Label       string              `json:"label"`
	Func        string              `json:"func"`
	Mode        string              `json:"mode"`
	Args        []string            `json:"args"`
	Opts        string              `json:"opts"`
	Lopts       string              `json:"lopts"`
	WantArgs    []string            `json:"want_args"`
	WantOptInd  int                 `json:"want_optind"`
	WantResults []fixtureRecordIter `json:"want_results"`
}

type fixture struct {
	Label       string
	Function    GetOptFunc
	Mode        GetOptMode
	Opts        []Opt
	LongOpts    []LongOpt
	Args        []string
	WantArgs    []string
	WantResults []fixtureIter
	WantOptInd  int
}

func (fr *fixtureRecord) toFixture() (fixture, error) {
	f := fixture{
		Label:      fr.Label,
		Args:       fr.Args,
		Opts:       OptStr(fr.Opts),
		LongOpts:   LongOptStr(fr.Lopts),
		WantArgs:   fr.WantArgs,
		WantOptInd: fr.WantOptInd,
	}

	switch fr.Func {
	case "getopt":
		f.Function = FuncGetOpt
	case "getopt_long":
		f.Function = FuncGetOptLong
	case "getopt_long_only":
		f.Function = FuncGetOptLongOnly
	default:
		return f, fmt.Errorf("unknown function type %q", fr.Func)
	}

	switch fr.Mode {
	case "gnu":
		f.Mode = ModeGNU
	case "posix":
		f.Mode = ModePosix
	case "inorder":
		f.Mode = ModeInOrder
	default:
		return f, fmt.Errorf("unknown mode type %q", fr.Mode)
	}

	for _, fri := range fr.WantResults {
		f.WantResults = append(f.WantResults, fri.toFixtureIter(&f))
	}

	return f, nil
}

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
		var record fixtureRecord
		if err := decoder.Decode(&record); err != nil {
			t.Fatalf("error decoding fixture: %v", err)
		}
		fixture, err := record.toFixture()
		if err != nil {
			t.Fatalf("error parsing fixture: %v", err)
		}
		testName := fmt.Sprintf("Fixture %q (function %q, mode %q)", record.Label, record.Func, record.Mode)
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

	if s.OptIndex != f.WantOptInd {
		t.Errorf("got OptIndex %d, but wanted %d", s.OptIndex, f.WantOptInd)
	}

	if !slices.Equal(s.Args, f.WantArgs) {
		t.Errorf("got Args %+q, but wanted %+q", s.Args, f.WantArgs)
	}
}
