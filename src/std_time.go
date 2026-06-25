package main

import (
	"fmt"
	"math"
	"math/big"
	"time"
)

const maxStdTimeSleepMilliseconds = math.MaxInt64 / int64(time.Millisecond)

func NewStdTimeMap() Value {
	entries := map[string]Binding{
		"CALENDAR_LOCAL":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdTimeLocalCalendar)),
		"CALENDAR_UTC":    NewImmutableBinding(NewBuiltinFunctionValue(builtinStdTimeUTCCalendar)),
		"SLEEP_MILLI":     NewImmutableBinding(NewBuiltinFunctionValue(builtinStdTimeMilliSleep)),
		"TIMESTAMP_MILLI": NewImmutableBinding(NewBuiltinFunctionValue(builtinStdTimeMilliTimestamp)),
		"TIMESTAMP_NANO":  NewImmutableBinding(NewBuiltinFunctionValue(builtinStdTimeNanoTimestamp)),
	}

	return NewMapValue(entries, true)
}

func builtinStdTimeLocalCalendar(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("TIME.CALENDAR_LOCAL expected 0 arguments, got %d", len(args))
	}

	return stdTimeCalendarSnapshot(time.Now()), nil
}

func builtinStdTimeUTCCalendar(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("TIME.CALENDAR_UTC expected 0 arguments, got %d", len(args))
	}

	return stdTimeCalendarSnapshot(time.Now().UTC()), nil
}

func builtinStdTimeMilliTimestamp(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("TIME.TIMESTAMP_MILLI expected 0 arguments, got %d", len(args))
	}

	return NewNumberValueFromInt64(time.Now().UnixMilli()), nil
}

func builtinStdTimeNanoTimestamp(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return Value{}, fmt.Errorf("TIME.TIMESTAMP_NANO expected 0 arguments, got %d", len(args))
	}

	return NewNumberValueFromInt64(time.Now().UnixNano()), nil
}

func builtinStdTimeMilliSleep(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return Value{}, fmt.Errorf("TIME.SLEEP_MILLI expected 1 argument, got %d", len(args))
	}

	milliseconds, err := stdTimeMillisecondsArg("TIME.SLEEP_MILLI", args[0], 1)
	if err != nil {
		return Value{}, err
	}

	time.Sleep(time.Duration(milliseconds) * time.Millisecond)

	return NewVoidValue(), nil
}

func stdTimeCalendarSnapshot(value time.Time) Value {
	zone, offset := value.Zone()

	entries := map[string]Binding{
		"DAY":        NewImmutableBinding(NewNumberValueFromInt(value.Day())),
		"HOUR":       NewImmutableBinding(NewNumberValueFromInt(value.Hour())),
		"MINUTE":     NewImmutableBinding(NewNumberValueFromInt(value.Minute())),
		"MONTH":      NewImmutableBinding(NewNumberValueFromInt(int(value.Month()))),
		"NANOSECOND": NewImmutableBinding(NewNumberValueFromInt(value.Nanosecond())),
		"OFFSET":     NewImmutableBinding(NewNumberValueFromInt(offset)),
		"SECOND":     NewImmutableBinding(NewNumberValueFromInt(value.Second())),
		"WEEKDAY":    NewImmutableBinding(NewNumberValueFromInt(int(value.Weekday()))),
		"YEAR":       NewImmutableBinding(NewNumberValueFromInt(value.Year())),
		"YEARDAY":    NewImmutableBinding(NewNumberValueFromInt(value.YearDay())),
		"ZONE":       NewImmutableBinding(NewStringValue(zone)),
	}

	return NewMapValue(entries, true)
}

func stdTimeMillisecondsArg(name string, arg Value, position int) (int64, error) {
	value := resolveSpecializedValue(arg)

	if value.Kind != ValueNumber {
		return 0, fmt.Errorf("%s argument %d expected a number of milliseconds", name, position)
	}

	milliseconds, accuracy := value.Number.Int64()
	if accuracy != big.Exact {
		return 0, fmt.Errorf("%s argument %d expected an integer number of milliseconds", name, position)
	}

	if milliseconds < 0 {
		return 0, fmt.Errorf("%s argument %d cannot be negative", name, position)
	}

	if milliseconds > maxStdTimeSleepMilliseconds {
		return 0, fmt.Errorf("%s argument %d is too large", name, position)
	}

	return milliseconds, nil
}
