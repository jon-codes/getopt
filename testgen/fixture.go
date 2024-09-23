package testgen

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"

	opt "github.com/jon-codes/getopt"
)

type caseRecord struct {
	label         string
	args          []string
	optstring     string
	longoptstring string
}

type fixtureParams struct {
	caseRecord
	step     int
	function opt.GetOptFunc
	mode     opt.GetOptMode
}

var inColNames = []string{"label", "args", "optstring", "longopts"}
var outColNames = []string{"label", "function", "mode", "step", "args", "optstring", "longopts", "char", "name", "err", "optindex", "optarg", "mutargs"}

var modes = []opt.GetOptMode{opt.ModeGNU, opt.ModePosix, opt.ModeInOrder}
var functions = []opt.GetOptFunc{opt.FuncGetOpt, opt.FuncGetOptLong, opt.FuncGetOptLongOnly}

func parseCaseRecord(row []string, cols map[string]int) (caseRecord, error) {
	var args []string

	err := json.Unmarshal([]byte(row[cols["args"]]), &args)
	if err != nil {
		return caseRecord{}, fmt.Errorf("error parsing JSON args array: %v", err)
	}

	label := row[cols["label"]]
	optstring := row[cols["optstring"]]
	longoptstring := row[cols["longopts"]]

	return caseRecord{
		label:         label,
		args:          args,
		optstring:     optstring,
		longoptstring: longoptstring,
	}, nil
}

func serializeFixtureResult(p fixtureParams, res cGetOptResult) ([]string, error) {
	record := []string{}

	jsonArgs, err := json.Marshal(p.args)
	if err != nil {
		return record, fmt.Errorf("error encoding JSON args: %v", err)
	}

	jsonMutArgs, err := json.Marshal(res.args)
	if err != nil {
		return record, fmt.Errorf("error encoding JSON args: %v", err)
	}

	record = append(record, p.label)
	record = append(record, p.function.String())
	record = append(record, p.mode.String())
	record = append(record, strconv.Itoa(p.step))
	record = append(record, string(jsonArgs))
	record = append(record, p.optstring)
	record = append(record, p.longoptstring)
	record = append(record, res.char)
	record = append(record, res.name)
	record = append(record, res.err)
	record = append(record, strconv.Itoa(res.optind))
	record = append(record, res.optarg)
	record = append(record, string(jsonMutArgs))

	return record, nil
}

func generateCaseFixtures(w *csv.Writer, record caseRecord) error {
	for _, function := range functions {
		for _, mode := range modes {
			step := 0

			cArgc, cArgv, free := BuildCArgv(record.args)

			for {
				p := fixtureParams{
					caseRecord: record,
					function:   function,
					mode:       mode,
					step:       step,
				}

				res, err := cGetOpt(cArgc, cArgv, p.optstring, p.longoptstring, function, mode)
				if err != nil {
					return fmt.Errorf("error generating case fixture: %v", err)
				}

				fixture, err := serializeFixtureResult(p, res)
				if err != nil {
					return err
				}

				err = w.Write(fixture)
				if err != nil {
					return fmt.Errorf("error writing outfile line: %v", err)
				}

				if res.err != "" {
					free()
					cResetGetOpt()
					break
				}

				step++
			}
		}
	}
	return nil
}

func ProcessCases(in *os.File, out *os.File) error {
	reader := csv.NewReader(in)
	writer := csv.NewWriter(out)
	defer writer.Flush()

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading infile header: %v", err)
	}

	inCols := map[string]int{}
	for _, colName := range inColNames {
		idx := slices.Index(header, colName)
		inCols[colName] = idx
	}

	err = writer.Write(outColNames)
	if err != nil {
		return fmt.Errorf("error writing outfile header: %v", err)
	}

	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading infile line: %v", err)
		}

		record, err := parseCaseRecord(row, inCols)
		if err != nil {
			return err
		}

		generateCaseFixtures(writer, record)
	}

	return nil
}
