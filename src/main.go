package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const expectedFileExtension = ".stult"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	switch len(args) {
	case 0:
		manifestPath, err := findManifestUpwards(".")
		if err != nil {
			return err
		}

		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			return err
		}

		return runManifest(manifest)

	case 1:
		target := args[0]

		manifestPath, isManifest, err := manifestPathFromArgument(target)
		if err != nil {
			return err
		}

		if isManifest {
			manifest, err := LoadManifest(manifestPath)
			if err != nil {
				return err
			}

			return runManifest(manifest)
		}

		interpreter := NewInterpreter()
		return runSourceFile(interpreter, target)

	default:
		return fmt.Errorf("Usage:\n  interpreter\n  interpreter <file.stul>\n  interpreter <directory>\n  interpreter <%s>", DefaultManifestFilename)
	}
}

func manifestPathFromArgument(target string) (string, bool, error) {
	info, err := os.Stat(target)

	if err == nil {
		if info.IsDir() {
			manifestPath, err := findManifestUpwards(target)
			if err != nil {
				return "", false, err
			}

			return manifestPath, true, nil
		}

		if isManifestFilename(target) {
			return target, true, nil
		}

		return "", false, nil
	}

	if os.IsNotExist(err) {
		if isManifestFilename(target) {
			return target, true, nil
		}

		return "", false, nil
	}

	return "", false, fmt.Errorf("Could not inspect %q: %w", target, err)
}

func findManifestUpwards(startDir string) (string, error) {
	absoluteDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("Could not resolve directory %q: %w", startDir, err)
	}

	info, err := os.Stat(absoluteDir)
	if err != nil {
		return "", fmt.Errorf("Could not inspect directory %q: %w", startDir, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("Expected %q to be a directory", startDir)
	}

	dir := absoluteDir

	for {
		manifestPath := filepath.Join(dir, DefaultManifestFilename)

		if _, err := os.Stat(manifestPath); err == nil {
			return manifestPath, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("Could not inspect manifest %q: %w", manifestPath, err)
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("Could not find %s in %q or any parent directory", DefaultManifestFilename, absoluteDir)
}

func isManifestFilename(filename string) bool {
	return filepath.Base(filename) == DefaultManifestFilename || filepath.Ext(filename) == ".json"
}

func runManifest(manifest *Manifest) error {
	interpreter := NewInterpreter()

	if err := applyManifestAliases(interpreter, manifest); err != nil {
		return err
	}

	for _, filename := range manifest.RunFiles {
		if err := runSourceFile(interpreter, filename); err != nil {
			return err
		}
	}

	return nil
}

func applyManifestAliases(interpreter *Interpreter, manifest *Manifest) error {
	for _, name := range manifest.AliasNames {
		expressionSource := manifest.Alias[name]

		value, err := evalManifestAliasExpression(interpreter, name, expressionSource)
		if err != nil {
			return err
		}

		if err := interpreter.Env.Set(name, value, isImmutableIdentifier(name)); err != nil {
			return fmt.Errorf("Could not set manifest alias %q: %w", name, err)
		}
	}

	return nil
}

func evalManifestAliasExpression(interpreter *Interpreter, name string, source string) (Value, error) {
	expression, err := parseManifestAliasExpression(name, source)
	if err != nil {
		return Value{}, err
	}

	value, err := interpreter.evalExpression(expression)
	if err != nil {
		return Value{}, fmt.Errorf("Could not evaluate manifest alias %q: %w", name, err)
	}

	return value, nil
}

func parseManifestAliasExpression(name string, source string) (Expression, error) {
	lexer := NewLexer(source)
	parser := NewParser(lexer)

	expression := parser.parseExpression(precLowest)

	if expression != nil {
		parser.skipSeparators()

		if parser.current.Type != TokenEOF {
			parser.errorAtCurrent("expected end of alias expression")
		}
	}

	if len(parser.Errors()) > 0 {
		return nil, formatManifestAliasParserErrors(name, parser.Errors())
	}

	if expression == nil {
		return nil, fmt.Errorf("Could not parse manifest alias %q", name)
	}

	return expression, nil
}

func runSourceFile(interpreter *Interpreter, filename string) error {
	if filepath.Ext(filename) != expectedFileExtension {
		fmt.Fprintf(
			os.Stderr,
			"Warning: Expected %s file extension, got %q\n",
			expectedFileExtension,
			filepath.Ext(filename),
		)
	}

	sourceBytes, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Could not read %q: %w", filename, err)
	}

	lexer := NewLexer(string(sourceBytes))
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return formatParserErrors(filename, parser.Errors())
	}

	if err := interpreter.EvalProgram(program); err != nil {
		return fmt.Errorf("Runtime error in %q: %w", filename, err)
	}

	return nil
}

func formatParserErrors(filename string, errors []string) error {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Parser errors in %q:", filename)

	for _, err := range errors {
		fmt.Fprintf(&builder, "\n  - %s", err)
	}

	return fmt.Errorf("%s", builder.String())
}

func formatManifestAliasParserErrors(name string, errors []string) error {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Parser errors in manifest alias %q:", name)

	for _, err := range errors {
		fmt.Fprintf(&builder, "\n  - %s", err)
	}

	return fmt.Errorf("%s", builder.String())
}
