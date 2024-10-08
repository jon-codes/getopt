package testgen

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type caseRecord struct {
	Label      string `json:"label"`
	ArgsStr    string `json:"args"`
	OptStr     string `json:"opts"`
	LongOptStr string `json:"longopts"`
}

func (c caseRecord) Args() (args []string) {
	return argsStr(c.ArgsStr)
}

var modes = []string{"gnu", "posix", "inorder"}
var functions = []string{"getopt", "getopt_long", "getopt_long_only"}

type fixtureIter struct {
	CharStr string `json:"char"`
	Name    string `json:"name"`
	OptArg  string `json:"opt_arg"`
	ErrStr  string `json:"err"`
}

type FixtureRecord struct {
	Label        string        `json:"label"`
	ArgsStr      string        `json:"args"`
	OptStr       string        `json:"opts"`
	LongOptStr   string        `json:"longopts"`
	FunctionStr  string        `json:"function"`
	ModeStr      string        `json:"mode"`
	WantArgsStr  string        `json:"want_args"`
	WantOptIndex int           `json:"want_optindex"`
	WantResults  []fixtureIter `json:"want_results"`
}

func generateCaseFixtures(w io.Writer, c caseRecord, more bool) (err error) {
	for i, function := range functions {
		for j, mode := range modes {
			step := 0

			f := FixtureRecord{
				Label:       c.Label,
				ArgsStr:     c.ArgsStr,
				OptStr:      c.OptStr,
				LongOptStr:  c.LongOptStr,
				FunctionStr: function,
				ModeStr:     mode,
			}
			cArgc, cArgv, free := BuildCArgv(c.Args())

			for {
				res, err := cGetOpt(cArgc, cArgv, c.OptStr, c.LongOptStr, function, mode)
				if err != nil {
					return fmt.Errorf("error generating case fixture: %v", err)
				}

				f.WantResults = append(f.WantResults, fixtureIter{
					CharStr: res.char,
					Name:    res.name,
					OptArg:  res.optarg,
					ErrStr:  res.err,
				})

				if res.err != "" {
					// save the final args & optindex
					f.WantArgsStr = strings.Join(res.args, " ")
					f.WantOptIndex = res.optind

					free()
					cResetGetOpt()
					break
				}

				step++
			}

			data, err := json.MarshalIndent(f, "\t", "\t")
			if err != nil {
				return fmt.Errorf("error marshalling case fixture: %v", err)
			}
			w.Write(data)

			if !(i == len(functions)-1 && j == len(modes)-1) {
				w.Write([]byte(",\n\t"))
			}
		}
	}

	if more {
		w.Write([]byte(",\n\t"))
	}

	return nil
}

func ProcessCases(in *os.File, out *os.File) error {
	decoder := json.NewDecoder(in)
	out.WriteString("[\n\t")

	// read open bracket
	_, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding cases: %v", err)
	}

	// while the array contains values
	for decoder.More() {
		var c caseRecord
		if err := decoder.Decode(&c); err != nil {
			return fmt.Errorf("error decoding cases: %v", err)
		}
		generateCaseFixtures(out, c, decoder.More())
	}

	// read closing bracket
	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding cases: %v", err)
	}

	out.WriteString("]")

	return nil
}

func argsStr(argsStr string) (args []string) {
	// TODO: this parsing is extremely basic and should be expanded to include scenarios like https://github.com/google/shlex

	var current strings.Builder
	inSingle := false
	inDouble := false

	for _, r := range argsStr {
		switch {
		case r == '"' && !inSingle:
			current.WriteRune(r)
		case r == '\'' && !inDouble:
			current.WriteRune(r)
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
