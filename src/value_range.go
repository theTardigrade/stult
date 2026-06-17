package main

import (
	"fmt"
	"math/big"
)

type stultRangeIterator struct {
	next        *big.Int
	end         *big.Int
	step        *big.Int
	ascending   bool
	isInclusive bool
}

func newStultRangeIterator(
	startValue Value,
	endValue Value,
	stepValue Value,
	isInclusive bool,
) (*stultRangeIterator, error) {
	start, err := numberToExactInteger(startValue, "range start")
	if err != nil {
		return nil, err
	}

	end, err := numberToExactInteger(endValue, "range end")
	if err != nil {
		return nil, err
	}

	step := big.NewInt(1)
	stepValue = resolveSpecializedValue(stepValue)
	if stepValue.Kind != ValueVoid {
		step, err = numberToExactInteger(stepValue, "range step")
		if err != nil {
			return nil, err
		}

		if step.Sign() <= 0 {
			return nil, fmt.Errorf("range step must be positive")
		}
	}

	return &stultRangeIterator{
		next:        start,
		end:         end,
		step:        step,
		ascending:   start.Cmp(end) <= 0,
		isInclusive: isInclusive,
	}, nil
}

func (iterator *stultRangeIterator) nextValue() (Value, bool) {
	if iterator == nil || iterator.next == nil || iterator.end == nil || iterator.step == nil {
		return NewVoidValue(), false
	}

	comparison := iterator.next.Cmp(iterator.end)
	if iterator.ascending {
		if iterator.isInclusive {
			if comparison > 0 {
				return NewVoidValue(), false
			}
		} else if comparison >= 0 {
			return NewVoidValue(), false
		}
	} else {
		if iterator.isInclusive {
			if comparison < 0 {
				return NewVoidValue(), false
			}
		} else if comparison <= 0 {
			return NewVoidValue(), false
		}
	}

	current := new(big.Int).Set(iterator.next)
	if iterator.ascending {
		iterator.next.Add(iterator.next, iterator.step)
	} else {
		iterator.next.Sub(iterator.next, iterator.step)
	}

	return NewNumberValueFromBigInt(current), true
}

func stultRangeValues(
	startValue Value,
	endValue Value,
	stepValue Value,
	isInclusive bool,
) ([]Value, error) {
	iterator, err := newStultRangeIterator(startValue, endValue, stepValue, isInclusive)
	if err != nil {
		return nil, err
	}

	values := []Value{}
	for {
		value, ok := iterator.nextValue()
		if !ok {
			break
		}

		values = append(values, value)
	}

	return values, nil
}

func numberToExactInteger(value Value, name string) (*big.Int, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueNumber {
		return nil, fmt.Errorf("%s must be a number", name)
	}

	out, accuracy := value.Number.Int(nil)
	if accuracy != big.Exact {
		return nil, fmt.Errorf("%s must be an integer", name)
	}

	return out, nil
}
