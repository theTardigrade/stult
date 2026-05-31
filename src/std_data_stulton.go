package main

import (
	"fmt"
	"sort"
	"strings"
)

func NewStdDataStultonMap() Value {
	entries := map[string]Binding{
		"IS_VALID": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdDataStultonIsValid)),
		"NEW":      NewImmutableBinding(NewBuiltinFunctionValue(builtinStdDataStultonNew)),
		"PARSE":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdDataStultonParse)),
	}

	return NewMapValue(entries, true)
}

func builtinStdDataStultonNew(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("DATA.STULTON.NEW expected 1 argument, got %d", len(args))
	}

	text, err := stdDataStultonFormatValue(args[0], 0)
	if err != nil {
		return Value{}, err
	}

	return NewStringValue(text), nil
}

func builtinStdDataStultonParse(_ *RuntimeContext, args []Value) (Value, error) {
	text, err := stdDataStultonStringArg("DATA.STULTON.PARSE", args)
	if err != nil {
		return Value{}, err
	}

	return stdDataStultonParseText(text)
}

func builtinStdDataStultonIsValid(_ *RuntimeContext, args []Value) (Value, error) {
	text, err := stdDataStultonStringArg("DATA.STULTON.IS_VALID", args)
	if err != nil {
		return Value{}, err
	}

	_, err = stdDataStultonParseText(text)
	return NewBoolValue(err == nil), nil
}

func stdDataStultonStringArg(name string, args []Value) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%s expected 1 argument, got %d", name, len(args))
	}

	value := resolveSpecializedValue(args[0])

	if value.Kind != ValueString {
		return "", fmt.Errorf("%s expected a string", name)
	}

	if value.Text == nil {
		return "", fmt.Errorf("%s expected a valid string", name)
	}

	return value.Text.String(), nil
}

func stdDataStultonParseText(text string) (Value, error) {
	lexer := NewLexer(text)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return Value{}, fmt.Errorf("DATA.STULTON.PARSE invalid STULTON: %s", strings.Join(parser.Errors(), "; "))
	}

	if len(program.Statements) != 1 {
		return Value{}, fmt.Errorf("DATA.STULTON.PARSE expected exactly one data value")
	}

	statement, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		return Value{}, fmt.Errorf("DATA.STULTON.PARSE expected a data expression, got executable syntax")
	}

	return stdDataStultonValueFromExpression(statement.Expression)
}

func stdDataStultonValueFromExpression(expr Expression) (Value, error) {
	switch e := expr.(type) {
	case *VoidLiteral:
		return NewVoidValue(), nil

	case *BoolLiteral:
		return NewBoolValue(e.Value), nil

	case *NumberLiteral:
		return stdDataStultonNumberValue(e.Value)

	case *StringLiteral:
		return NewStringValue(e.Value), nil

	case *PrefixExpression:
		if e.Operator == "-" {
			number, ok := e.Right.(*NumberLiteral)
			if !ok {
				return Value{}, fmt.Errorf("DATA.STULTON.PARSE only allows unary '-' before plain numbers")
			}

			return stdDataStultonNumberValue("-" + number.Value)
		}

		return Value{}, fmt.Errorf("DATA.STULTON.PARSE does not allow prefix operator %q", e.Operator)

	case *ArrayLiteral:
		elements := make([]Value, 0, len(e.Elements))

		for _, element := range e.Elements {
			expressionElement, ok := element.(*ExpressionArrayElement)
			if !ok {
				return Value{}, fmt.Errorf("DATA.STULTON.PARSE does not allow ranges in data arrays")
			}

			value, err := stdDataStultonValueFromExpression(expressionElement.Expression)
			if err != nil {
				return Value{}, err
			}

			elements = append(elements, value)
		}

		return NewArrayValue(elements, false), nil

	case *MapLiteral:
		entries := make(map[string]Binding)

		for _, entry := range e.Entries {
			key := entry.Key.Literal

			if _, exists := entries[key]; exists {
				return Value{}, fmt.Errorf(
					"line %d, column %d: DATA.STULTON.PARSE duplicate map key %q",
					entry.Key.StartOfLine,
					entry.Key.StartOfColumn,
					key,
				)
			}

			value, err := stdDataStultonValueFromExpression(entry.Value)
			if err != nil {
				return Value{}, err
			}

			entries[key] = Binding{
				Value:       value,
				IsImmutable: isImmutableIdentifier(key),
			}
		}

		return NewMapValue(entries, false), nil

	default:
		return Value{}, fmt.Errorf("DATA.STULTON.PARSE does not allow %s", stdDataStultonExpressionName(expr))
	}
}

func stdDataStultonNumberValue(literal string) (Value, error) {
	if strings.ContainsAny(literal, "eE") {
		return Value{}, fmt.Errorf("DATA.STULTON.PARSE does not allow exponential number notation")
	}

	return NewNumberValueFromString(literal)
}

func stdDataStultonExpressionName(expr Expression) string {
	switch expr.(type) {
	case *IdentifierExpression:
		return "identifiers"

	case *BinaryExpression:
		return "operators"

	case *IndexExpression:
		return "index expressions"

	case *FunctionLiteral:
		return "function literals"

	case *CallExpression:
		return "function calls"

	default:
		return fmt.Sprintf("expression type %T", expr)
	}
}

func stdDataStultonFormatValue(value Value, indent int) (string, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueVoid:
		return "_", nil

	case ValueBool:
		if value.Bool {
			return "\\/", nil
		}

		return "/\\", nil

	case ValueNumber:
		if value.Number == nil {
			return "", fmt.Errorf("DATA.STULTON.NEW cannot encode invalid number")
		}

		return formatNumber(value.Number, DefaultFractionDigits), nil

	case ValueString:
		if value.Text == nil {
			return "", fmt.Errorf("DATA.STULTON.NEW cannot encode invalid string")
		}

		return stdDataStultonQuoteString(value.Text.String())

	case ValueArray:
		return stdDataStultonFormatArray(value.Array, indent)

	case ValueMap:
		return stdDataStultonFormatMap(value.Map, indent)

	case ValueFunction, ValueBuiltinFunction:
		return "", fmt.Errorf("DATA.STULTON.NEW cannot encode functions")

	default:
		return "", fmt.Errorf("DATA.STULTON.NEW cannot encode unknown value kind")
	}
}

func stdDataStultonFormatArray(array *Array, indent int) (string, error) {
	if array == nil {
		return "", fmt.Errorf("DATA.STULTON.NEW cannot encode invalid array")
	}

	if len(array.Elements) == 0 {
		return "{}", nil
	}

	currentIndent := strings.Repeat("\t", indent)
	childIndent := strings.Repeat("\t", indent+1)

	lines := []string{"{"}

	for _, element := range array.Elements {
		text, err := stdDataStultonFormatValue(element, indent+1)
		if err != nil {
			return "", err
		}

		lines = append(lines, childIndent+text)
	}

	lines = append(lines, currentIndent+"}")

	return strings.Join(lines, "\n"), nil
}

func stdDataStultonFormatMap(m *Map, indent int) (string, error) {
	if m == nil {
		return "", fmt.Errorf("DATA.STULTON.NEW cannot encode invalid map")
	}

	if len(m.Entries) == 0 {
		return "{:}", nil
	}

	currentIndent := strings.Repeat("\t", indent)
	childIndent := strings.Repeat("\t", indent+1)

	keys := make([]string, 0, len(m.Entries))
	for key := range m.Entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := []string{"{"}

	for _, key := range keys {
		quotedKey, err := stdDataStultonQuoteString(key)
		if err != nil {
			return "", err
		}

		text, err := stdDataStultonFormatValue(m.Entries[key].Value, indent+1)
		if err != nil {
			return "", err
		}

		lines = append(lines, childIndent+quotedKey+": "+text)
	}

	lines = append(lines, currentIndent+"}")

	return strings.Join(lines, "\n"), nil
}

func stdDataStultonQuoteString(text string) (string, error) {
	var out strings.Builder

	out.WriteRune('"')

	for _, ch := range text {
		switch ch {
		case '\n':
			out.WriteString(`\n`)

		case '\t':
			out.WriteString(`\t`)

		case '"':
			out.WriteString(`\"`)

		case '\\':
			out.WriteString(`\\`)

		default:
			if ch < 0x20 {
				return "", fmt.Errorf("DATA.STULTON.NEW cannot encode unsupported string control character U+%04X", ch)
			}

			out.WriteRune(ch)
		}
	}

	out.WriteRune('"')

	return out.String(), nil
}
