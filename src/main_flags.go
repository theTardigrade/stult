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

type DumpCommandOptions struct {
	OutputPath string
}

func parseDumpArgs(args []string) ([]string, DumpCommandOptions, error) {
	var dumpArgs []string
	var options DumpCommandOptions

	for index := 0; index < len(args); index++ {
		arg := args[index]

		switch arg {
		case "--bytecode":
			// `dump` is bytecode-only. The flag is accepted for symmetry with `run`.

		case "-o", "--output":
			if index+1 >= len(args) {
				return nil, options, fmt.Errorf("dump option %s requires an output path", arg)
			}

			options.OutputPath = args[index+1]
			index++

		default:
			if isEvalFlag(arg) {
				dumpArgs = append(dumpArgs, arg)

				if index+1 >= len(args) {
					return nil, options, fmt.Errorf("%s requires a source string", arg)
				}

				index++
				dumpArgs = append(dumpArgs, args[index])

				continue
			}

			if strings.HasPrefix(arg, "-") && !isStdinTarget(arg) {
				return nil, options, fmt.Errorf("unknown dump option %q\n%s", arg, dumpUsage())
			}

			dumpArgs = append(dumpArgs, arg)
		}
	}

	return dumpArgs, options, nil
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
		"  stult dump [--bytecode] [-o|--output <output-file>] [file.stult|directory|manifest|-]\n" +
		"  stult dump [--bytecode] [-o|--output <output-file>] -e|--eval <source-string>\n" +
		"  stult build [--bytecode|--interpreter] [project-directory-or-file.stult] -o <output-executable>"
}

func dumpUsage() string {
	return "Usage:\n" +
		"  stult dump [--bytecode] [-o|--output <output-file>] [file.stult|directory|manifest|-]\n" +
		"  stult dump [--bytecode] [-o|--output <output-file>] -e|--eval <source-string>"
}

func printUsage() {
	fmt.Println(commandUsage())
}
