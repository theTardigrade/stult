package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

func NewStdCSVMap() Value {
	entries := map[string]Binding{
		"IS_VALID": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdCSVIsValid)),
		"NEW":      NewImmutableBinding(NewBuiltinFunctionValue(builtinStdCSVNew)),
		"PARSE":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdCSVParse)),
	}

	return NewMapValue(entries, true)
}

func builtinStdCSVNew(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("CSV.NEW expected 1 argument, got %d", len(args))
	}

	rows, err := stdCSVRowsFromValue(args[0])
	if err != nil {
		return Value{}, err
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return Value{}, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return Value{}, err
	}

	return NewStringValue(builder.String()), nil
}

func builtinStdCSVParse(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("CSV.PARSE expected 1 argument, got %d", len(args))
	}

	text, err := stdCSVStringArg("CSV.PARSE", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	reader := stdCSVReader(text)

	records, err := reader.ReadAll()
	if err != nil {
		return Value{}, err
	}

	return stdCSVRecordsToValue(records), nil
}

func builtinStdCSVIsValid(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("CSV.IS_VALID expected 1 argument, got %d", len(args))
	}

	text, err := stdCSVStringArg("CSV.IS_VALID", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	reader := stdCSVReader(text)

	for {
		_, err := reader.Read()
		if err == nil {
			continue
		}

		if err == io.EOF {
			return NewBoolValue(true), nil
		}

		return NewBoolValue(false), nil
	}
}

func stdCSVReader(text string) *csv.Reader {
	reader := csv.NewReader(strings.NewReader(text))

	// CSV data in the wild often has records with different field counts.
	// Let users validate shape themselves if they need fixed-width rows.
	reader.FieldsPerRecord = -1

	return reader
}

func stdCSVStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string", name, position)
	}

	return value.Text.String(), nil
}

func stdCSVRecordsToValue(records [][]string) Value {
	rows := make([]Value, 0, len(records))

	for _, record := range records {
		fields := make([]Value, 0, len(record))

		for _, field := range record {
			fields = append(fields, NewStringValue(field))
		}

		rows = append(rows, NewArrayValue(fields, false))
	}

	return NewArrayValue(rows, false)
}

func stdCSVRowsFromValue(value Value) ([][]string, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueArray || value.Array == nil {
		return nil, fmt.Errorf("CSV.NEW expected an array of row arrays")
	}

	rows := make([][]string, 0, len(value.Array.Elements))

	if err := value.Array.ForEach(func(rowIndex *Number, rowValue Value) error {
		row, err := stdCSVRowFromValue(rowValue, rowIndex)
		if err != nil {
			return err
		}

		rows = append(rows, row)
		return nil
	}); err != nil {
		return nil, err
	}

	return rows, nil
}

func stdCSVRowFromValue(value Value, rowIndex *Number) ([]string, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueArray || value.Array == nil {
		return nil, fmt.Errorf("CSV.NEW row %s expected an array", formatArrayIndex(rowIndex))
	}

	row := make([]string, 0, len(value.Array.Elements))

	if err := value.Array.ForEach(func(columnIndex *Number, fieldValue Value) error {
		field, err := stdCSVFieldFromValue(fieldValue)
		if err != nil {
			return fmt.Errorf("CSV.NEW row %s column %s: %w",
				formatArrayIndex(rowIndex),
				formatArrayIndex(columnIndex),
				err,
			)
		}

		row = append(row, field)
		return nil
	}); err != nil {
		return nil, err
	}

	return row, nil
}

func stdCSVFieldFromValue(value Value) (string, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueString:
		if value.Text == nil {
			return "", fmt.Errorf("invalid string")
		}

		return value.Text.String(), nil

	case ValueVoid,
		ValueNumber,
		ValueBool,
		ValueMap,
		ValueArray,
		ValueFunction,
		ValueBuiltinFunction:
		return value.PrintString(), nil

	default:
		return "", fmt.Errorf("unsupported field type")
	}
}
