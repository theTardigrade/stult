package main

import (
	"fmt"
	"strings"
)

func parseRuntimeModeFlag(args []string) (RuntimeMode, []string, error) {
	if len(args) == 0 {
		return RuntimeModeBytecode, args, nil
	}

	switch args[0] {
	case "--bytecode":
		return RuntimeModeBytecode, args[1:], nil

	case "--interpreter":
		return RuntimeModeInterpreter, args[1:], nil

	default:
		return RuntimeModeBytecode, args, nil
	}
}

func parseDumpArgs(args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

	switch args[0] {
	case "--bytecode":
		return args[1:], nil

	default:
		if strings.HasPrefix(args[0], "-") && !isEvalFlag(args[0]) && !isStdinTarget(args[0]) {
			return nil, fmt.Errorf("unknown dump option %q\n%s", args[0], dumpUsage())
		}

		return args, nil
	}
}

func isEvalFlag(arg string) bool {
	return arg == "-e" || arg == "--eval"
}

func isStdinTarget(arg string) bool {
	return arg == "-"
}

func commandUsage() string {
	return "Usage:\n" +
		"  stult run [--bytecode|--interpreter] [file.stult|directory|manifest|-] [args...]\n" +
		"  stult run [--bytecode|--interpreter] -e|--eval <source-string> [args...]\n" +
		"  stult dump [--bytecode] [file.stult|directory|manifest|-]\n" +
		"  stult dump [--bytecode] -e|--eval <source-string>\n" +
		"  stult build [--bytecode|--interpreter] [project-directory-or-file.stult] -o <output-executable>"
}

func dumpUsage() string {
	return "Usage:\n" +
		"  stult dump [--bytecode] [file.stult|directory|manifest|-]\n" +
		"  stult dump [--bytecode] -e|--eval <source-string>"
}

func printUsage() {
	fmt.Println(commandUsage())
}
