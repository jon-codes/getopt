package testgen

/*
#include <stdlib.h>
#include <getopt.h>

char* get_optarg() {
	return optarg;
}

int get_optind() {
	return optind;
}

int get_optopt() {
	return optopt;
}

int reset_getopt() {
	optind = 0;
}
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

type cGetOptResult struct {
	char   string
	name   string
	err    string
	optind int
	optarg string
	args   []string
}

func BuildCArgv(args []string) (C.int, []*C.char, func()) {
	cArgc := C.int(len(args))
	cArgv := make([]*C.char, cArgc)
	for i, arg := range args {
		cArgv[i] = C.CString(arg)
	}
	free := func() {
		for i := 0; i < len(args); i++ {
			C.free(unsafe.Pointer(cArgv[i]))
		}
	}
	return cArgc, cArgv, free
}

func buildCOptstring(optstring string, mode string) (*C.char, func()) {
	optstring = ":" + optstring // always act like opterr = 0

	if mode == "posix" { // posixly_correct
		optstring = "+" + optstring
	}
	if mode == "inorder" { // inorder
		optstring = "-" + optstring
	}
	cOptstring := C.CString(optstring)
	free := func() {
		C.free(unsafe.Pointer(cOptstring))
	}
	return cOptstring, free
}

func buildCLongoptions(longoptstring string, flag *C.int) ([]C.struct_option, func()) {
	longoptions := []C.struct_option{}

	opts := strings.Split(longoptstring, ",")
	for idx, opt := range opts {
		name := strings.TrimSpace(opt)
		if name == "" {
			continue
		}
		hasArg := 0 // no_argument
		name, found := strings.CutSuffix(opt, "::")
		if found {
			hasArg = 2 // optional_argument
		} else {
			name, found = strings.CutSuffix(opt, ":")
			if found {
				hasArg = 1 // required_argument
			}
		}

		longoptions = append(longoptions, C.struct_option{
			name:    C.CString(name),
			has_arg: C.int(hasArg),
			flag:    flag,
			val:     C.int(-(idx + 1)),
		})
	}
	// null terminator
	longoptions = append(longoptions, C.struct_option{name: nil, has_arg: 0, flag: nil, val: 0})

	free := func() {
		for _, opt := range longoptions {
			if opt.name != nil {
				C.free(unsafe.Pointer(opt.name))
			}
		}
	}
	return longoptions, free
}

func parseRet(ret int, optopt int) (string, string) {
	char := ""
	err := ""
	if ret == -1 {
		err = "-1"
	} else if ret == ':' || ret == '?' {
		err = string(rune(ret))
		if optopt > 0 {
			char = string(rune(optopt))
		}
	} else if ret != 0 {
		char = string(rune(ret))
	}
	return char, err
}

func parseName(cLongoptions []C.struct_option, char string, optopt int, flag *C.int) string {
	name := ""
	if char != "" && char != ":" && char != "?" {
		return name
	}
	if optopt < 0 {
		name = C.GoString(cLongoptions[(-(optopt) - 1)].name)
	}
	if *flag < 0 {
		name = C.GoString(cLongoptions[(-(*flag) - 1)].name)
	}
	return name
}

func copyCArgv(cArgc C.int, cArgv []*C.char) []string {
	args := make([]string, cArgc)
	for i := 0; i < int(cArgc); i++ {
		args[i] = C.GoString(cArgv[i])
	}
	return args
}

func cGetOpt(cArgc C.int, cArgv []*C.char, optstring string, longoptstring string, function string, mode string) (cGetOptResult, error) {
	cOptstring, freeCOptstring := buildCOptstring(optstring, mode)
	defer freeCOptstring()

	flag := (*C.int)(C.malloc(C.sizeof_int))
	*flag = 0
	defer C.free(unsafe.Pointer(flag))

	cLongoptions, freeCLongoptions := buildCLongoptions(longoptstring, flag)
	defer freeCLongoptions()

	var cLongindex C.int
	var ret int

	switch function {
	case "getopt":
		ret = int(C.getopt(cArgc, &cArgv[0], cOptstring))
	case "getopt_long":
		ret = int(C.getopt_long(cArgc, &cArgv[0], cOptstring, &cLongoptions[0], &cLongindex))
	case "getopt_long_only":
		ret = int(C.getopt_long_only(cArgc, &cArgv[0], cOptstring, &cLongoptions[0], &cLongindex))
	default:
		return cGetOptResult{}, fmt.Errorf("unknown function type: %q", function)
	}

	optarg := C.GoString(C.get_optarg())
	optind := int(C.get_optind())
	optopt := int(C.get_optopt())
	curr_arg := C.GoString(cArgv[optind-1])

	char, err := parseRet(ret, optopt)
	name := parseName(cLongoptions, char, optopt, flag)
	mutArgs := copyCArgv(cArgc, cArgv)

	if (err == "?" || err == ":") && (char == "" && name == "") {
		val := strings.TrimPrefix(curr_arg, "-")
		val = strings.TrimPrefix(val, "-")
		if len(val) == 1 {
			char = val
		} else {
			name = val
		}
	}

	return cGetOptResult{
		char:   char,
		name:   name,
		err:    err,
		optind: optind,
		optarg: optarg,
		args:   mutArgs,
	}, nil
}

func cResetGetOpt() {
	C.reset_getopt()
}
