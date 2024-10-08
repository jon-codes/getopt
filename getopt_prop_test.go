package getopt_test

import (
	"testing"

	"github.com/jon-codes/getopt"
	"pgregory.net/rapid"
)

var (
	genHasArg = rapid.SampledFrom([]getopt.HasArg{getopt.NoArgument, getopt.RequiredArgument, getopt.OptionalArgument})
	funcGen   = rapid.SampledFrom([]getopt.Func{getopt.FuncGetOpt, getopt.FuncGetOptLong, getopt.FuncGetOptLongOnly})
	modeGen   = rapid.SampledFrom([]getopt.Mode{getopt.ModeGNU, getopt.ModePosix, getopt.ModeInOrder})
)

var optGen = rapid.Custom(func(t *rapid.T) getopt.Opt {
	return getopt.Opt{Char: rapid.Rune().Draw(t, "char"), HasArg: genHasArg.Draw(t, "has_arg")}
})

var longOptGen = rapid.Custom(func(t *rapid.T) getopt.LongOpt {
	return getopt.LongOpt{Name: rapid.String().Draw(t, "name"), HasArg: genHasArg.Draw(t, "has_arg")}
})

var configGen = rapid.Custom(func(t *rapid.T) getopt.Config {
	return getopt.Config{
		Opts:     rapid.SliceOf(optGen).Draw(t, "opts"),
		LongOpts: rapid.SliceOf(longOptGen).Draw(t, "long_opts"),
		Func:     funcGen.Draw(t, "func"),
		Mode:     modeGen.Draw(t, "mode"),
	}
})

func propTarget(t *rapid.T) {
	args := rapid.SliceOfN(rapid.String(), 0, -1).Draw(t, "args")

	c := configGen.Draw(t, "config")
	s := getopt.NewState(args)

	prevOptInd := s.OptInd()
	for res, err := range s.All(c) {
		if err != nil && !(err == getopt.ErrMissingOptArg || err == getopt.ErrIllegalOptArg || err == getopt.ErrUnknownOpt) {
			t.Fatalf("unknown err returned: %v", err)
		}

		if err != nil {
			if err == getopt.ErrMissingOptArg && res.OptArg != "" {
				t.Fatalf("result has OptArg %q, but err claims it is missing", res.OptArg)
			}
		}

		if res.Char != 0 {
			if res.Name != "" {
				t.Fatalf("result has both Char %q and Name %q", res.Char, res.Name)
			}
		}

		if res.Name != "" {
			if res.Char != 0 {
				t.Fatalf("result has both Char %q and Name %q", res.Char, res.Name)
			}
		}

		if s.OptInd() < prevOptInd {
			t.Fatalf("OptInd decreased from %d to %d", prevOptInd, s.OptInd())
		}
	}

	if s.OptInd() > 1 && s.OptInd() > len(s.Args())+1 {
		t.Fatalf("OptInd exceeded last arg + 1: args len is %d, bug OptInd is %d", len(args), s.OptInd())
	}
}

func TestGetOpt_Property(t *testing.T) {
	rapid.Check(t, propTarget)
}

func FuzzGetOpt(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(propTarget))
}
