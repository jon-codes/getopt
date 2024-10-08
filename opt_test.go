package opt

import (
	"errors"
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	g := NewGetOptState([]string{"my_program", "-a", "-b"})
	assertState(t, g, GetOptState{Args: []string{"my_program", "-a", "-b"}, OptIndex: 1, ArgIndex: 0})
}

func TestReset(t *testing.T) {
	g := NewGetOptState([]string{"my_program", "-a", "-b"})
	p := GetOptParams{Short: []ShortOpt{{Char: 'a'}, {Char: 'b'}}}

	_, _ = g.GetOpt(p)
	assertState(t, g, GetOptState{Args: []string{"my_program", "-a", "-b"}, OptIndex: 2, ArgIndex: 0})

	g.Reset([]string{"my_program", "-b"})
	assertState(t, g, GetOptState{Args: []string{"my_program", "-b"}, OptIndex: 1, ArgIndex: 0})
}

func TestGetOpt_Opts(t *testing.T) {
	t.Run("a single valid option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a"})
		p := GetOptParams{Short: []ShortOpt{{Char: 'a'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: 'a'}, nil},
			{GetOptResult{}, ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a"}, 2)
	})

	t.Run("a single valid option with arguments", func(t *testing.T) {
		t.Run("when the argument is provided inline", func(t *testing.T) {
			g := NewGetOptState([]string{"my_program", "-afoo"})
			p := GetOptParams{Short: []ShortOpt{{Char: 'a', HasArg: RequiredArgument}}}

			wants := []resultAssertion{
				{GetOptResult{Char: 'a', OptArg: "foo"}, nil},
				{GetOptResult{}, ErrDone},
			}

			assertSequence(t, g, p, wants)
			assertArgs(t, g, []string{"my_program", "-afoo"}, 2)
		})

		t.Run("when the argument is provided in the next arg", func(t *testing.T) {
			g := NewGetOptState([]string{"my_program", "-a", "foo"})
			p := GetOptParams{Short: []ShortOpt{{Char: 'a', HasArg: RequiredArgument}}}

			wants := []resultAssertion{
				{GetOptResult{Char: 'a', OptArg: "foo"}, nil},
				{GetOptResult{}, ErrDone},
			}

			assertSequence(t, g, p, wants)
			assertArgs(t, g, []string{"my_program", "-a", "foo"}, 3)
		})

		t.Run("when the next arg looks like an option", func(t *testing.T) {
			g := NewGetOptState([]string{"my_program", "-a", "-b"})
			p := GetOptParams{Short: []ShortOpt{{Char: 'a', HasArg: RequiredArgument}}}

			wants := []resultAssertion{
				{GetOptResult{Char: 'a', OptArg: "-b"}, nil},
				{GetOptResult{}, ErrDone},
			}

			assertSequence(t, g, p, wants)
			assertArgs(t, g, []string{"my_program", "-a", "-b"}, 3)
		})

		t.Run("when the argument contains multi-byte chars", func(t *testing.T) {
			g := NewGetOptState([]string{"my_program", "-a文"})
			p := GetOptParams{Short: []ShortOpt{{Char: 'a', HasArg: RequiredArgument}}}

			wants := []resultAssertion{
				{GetOptResult{Char: 'a', OptArg: "文"}, nil},
				{GetOptResult{}, ErrDone},
			}

			assertSequence(t, g, p, wants)
			assertArgs(t, g, []string{"my_program", "-a文"}, 2)
		})
	})

	t.Run("multiple valid options", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "-b", "-c"})
		p := GetOptParams{Short: []ShortOpt{{Char: 'a'}, {Char: 'b'}, {Char: 'c'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: 'a'}, nil},
			{GetOptResult{Char: 'b'}, nil},
			{GetOptResult{Char: 'c'}, nil},
			{GetOptResult{}, ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "-b", "-c"}, 4)
	})

	t.Run("option groups", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-abc"})
		p := GetOptParams{Short: []ShortOpt{{Char: 'a'}, {Char: 'b'}, {Char: 'c'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: 'a'}, nil},
			{GetOptResult{Char: 'b'}, nil},
			{GetOptResult{Char: 'c'}, nil},
			{GetOptResult{}, ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-abc"}, 2)
	})

	t.Run("a single undefined option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-b"})
		p := GetOptParams{Short: []ShortOpt{{Char: 'a'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: 'b'}, ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-b"}, 1)
	})

	t.Run("a single '-' option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-"})
		p := GetOptParams{Short: []ShortOpt{{Char: '-'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: '-'}, ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-"}, 1)
	})

	t.Run("a single non-graph option", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-文"})
		p := GetOptParams{Short: []ShortOpt{{Char: '文'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: '文'}, ErrIllegalOpt},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-文"}, 1)
	})

	t.Run("-- terminates option parsing", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "--", "-b"})
		p := GetOptParams{Short: []ShortOpt{{Char: 'a'}, {Char: 'b'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: 'a'}, nil},
			{GetOptResult{}, ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "--", "-b"}, 3)
	})

	t.Run("positionals terminate option parsing", func(t *testing.T) {
		g := NewGetOptState([]string{"my_program", "-a", "my_arg", "-b"})
		p := GetOptParams{Short: []ShortOpt{{Char: 'a'}, {Char: 'b'}}}

		wants := []resultAssertion{
			{GetOptResult{Char: 'a'}, nil},
			{GetOptResult{}, ErrDone},
		}

		assertSequence(t, g, p, wants)
		assertArgs(t, g, []string{"my_program", "-a", "my_arg", "-b"}, 2)
	})
}

type resultAssertion struct {
	result GetOptResult
	err    error
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
		assertResult(t, got, want.result)
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
	if got.ArgIndex != want.ArgIndex {
		t.Errorf("got ArgIndex %d, but wanted %d", got.ArgIndex, want.ArgIndex)
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