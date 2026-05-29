package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var sourceLocationPattern = regexp.MustCompile(`\bline ([0-9]+), column ([0-9]+)`)

type sourceLocation struct {
	Line   int
	Column int
}

func formatParserErrors(filename string, source string, errors []string) error {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Parser errors in %q:", filename)

	for _, err := range errors {
		fmt.Fprintf(&builder, "\n  - %s", err)

		if location, ok := findSourceLocation(err); ok {
			appendSourceContext(&builder, source, location)
		}
	}

	return fmt.Errorf("%s", builder.String())
}

func formatRuntimeError(filename string, source string, err error) error {
	var builder strings.Builder

	message := err.Error()

	fmt.Fprintf(&builder, "Runtime error in %q:", filename)
	fmt.Fprintf(&builder, "\n  %s", message)

	if location, ok := findSourceLocation(message); ok {
		appendSourceContext(&builder, source, location)
	}

	return fmt.Errorf("%s", builder.String())
}

func findSourceLocation(message string) (sourceLocation, bool) {
	matches := sourceLocationPattern.FindStringSubmatch(message)
	if len(matches) != 3 {
		return sourceLocation{}, false
	}

	line, err := strconv.Atoi(matches[1])
	if err != nil {
		return sourceLocation{}, false
	}

	column, err := strconv.Atoi(matches[2])
	if err != nil {
		return sourceLocation{}, false
	}

	if line < 1 || column < 1 {
		return sourceLocation{}, false
	}

	return sourceLocation{
		Line:   line,
		Column: column,
	}, true
}

func appendSourceContext(builder *strings.Builder, source string, location sourceLocation) {
	line, ok := sourceLine(source, location.Line)
	if !ok {
		return
	}

	lineNumberWidth := len(strconv.Itoa(location.Line))
	gutterPadding := strings.Repeat(" ", lineNumberWidth)

	fmt.Fprintf(builder, "\n    %s |", gutterPadding)
	fmt.Fprintf(builder, "\n    %*d | %s", lineNumberWidth, location.Line, line)
	fmt.Fprintf(
		builder,
		"\n    %s | %s^",
		gutterPadding,
		caretPadding(line, location.Column),
	)
}

func sourceLine(source string, lineNumber int) (string, bool) {
	if lineNumber < 1 {
		return "", false
	}

	currentLine := 1
	lineStart := 0

	for index, ch := range source {
		if ch != '\n' {
			continue
		}

		if currentLine == lineNumber {
			return strings.TrimSuffix(source[lineStart:index], "\r"), true
		}

		currentLine++
		lineStart = index + 1
	}

	if currentLine == lineNumber {
		return strings.TrimSuffix(source[lineStart:], "\r"), true
	}

	return "", false
}

func caretPadding(line string, column int) string {
	if column <= 1 {
		return ""
	}

	var builder strings.Builder

	currentColumn := 1

	for _, ch := range line {
		if currentColumn >= column {
			break
		}

		if ch == '\t' {
			builder.WriteRune('\t')
		} else {
			builder.WriteRune(' ')
		}

		currentColumn++
	}

	for currentColumn < column {
		builder.WriteByte(' ')
		currentColumn++
	}

	return builder.String()
}
