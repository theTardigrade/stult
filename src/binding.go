package main

import "fmt"

type BindingContractKind int

const (
	BindingContractAnyKind BindingContractKind = iota
	BindingContractSameKind
	BindingContractExactKind
	BindingContractArrayKind
	BindingContractMapKind
	BindingContractUnionKind
)

type BindingContract struct {
	Kind         BindingContractKind
	KindValue    ValueKind
	HasKindValue bool
	Element      *BindingContract
	Options      []BindingContract
}

type BindingContractDeclaration struct {
	Token    Token
	Contract BindingContract
}

type Binding struct {
	Value       Value
	IsImmutable bool
	Contract    BindingContract
}

func NewImmutableBinding(value Value) Binding {
	return Binding{
		Value:       value,
		IsImmutable: true,
	}
}

func NewSameKindBindingContract(value Value) BindingContract {
	value = resolveSpecializedValue(value)

	return BindingContract{
		Kind:         BindingContractSameKind,
		KindValue:    value.Kind,
		HasKindValue: true,
	}
}

func BindingContractFromDeclaration(declaration *BindingContractDeclaration, initialValue Value) BindingContract {
	if declaration == nil {
		return BindingContract{}
	}

	_ = initialValue
	return declaration.Contract.Clone()
}

func (contract BindingContract) Clone() BindingContract {
	cloned := contract
	if contract.Element != nil {
		element := contract.Element.Clone()
		cloned.Element = &element
	}

	if contract.Options != nil {
		cloned.Options = make([]BindingContract, len(contract.Options))
		for index, option := range contract.Options {
			cloned.Options[index] = option.Clone()
		}
	}

	return cloned
}

func (contract BindingContract) ClonePointer() *BindingContract {
	cloned := contract.Clone()
	return &cloned
}

func (contract BindingContract) IsAny() bool {
	return contract.Kind == BindingContractAnyKind
}

func (contract BindingContract) Check(name string, value Value) error {
	checked := contract.Clone()
	return checked.CheckAndLearn(name, value)
}

func (contract *BindingContract) CheckAndLearn(name string, value Value) error {
	if contract == nil {
		return nil
	}

	value = resolveSpecializedValue(value)

	switch contract.Kind {
	case BindingContractAnyKind:
		return nil

	case BindingContractSameKind:
		if !contract.HasKindValue {
			contract.KindValue = value.Kind
			contract.HasKindValue = true
			return nil
		}

		if value.Kind != contract.KindValue {
			return fmt.Errorf(
				"binding %q expects %s value, got %s value",
				name,
				valueKindName(contract.KindValue),
				valueKindName(value.Kind),
			)
		}

		return nil

	case BindingContractExactKind:
		if value.Kind != contract.KindValue {
			return fmt.Errorf(
				"binding %q expects %s value, got %s value",
				name,
				valueKindName(contract.KindValue),
				valueKindName(value.Kind),
			)
		}

		return nil

	case BindingContractArrayKind:
		if value.Kind != ValueArray || value.Array == nil {
			return fmt.Errorf(
				"binding %q expects array value, got %s value",
				name,
				valueKindName(value.Kind),
			)
		}

		if contract.Element != nil {
			if err := checkArrayElementContract(name, value.Array, contract.Element); err != nil {
				return err
			}

			value.Array.ElementContract = contract.Element.ClonePointer()
		}

		return nil

	case BindingContractMapKind:
		if value.Kind != ValueMap || value.Map == nil {
			return fmt.Errorf(
				"binding %q expects map value, got %s value",
				name,
				valueKindName(value.Kind),
			)
		}

		if contract.Element != nil {
			if err := checkMapValueContract(name, value.Map, contract.Element); err != nil {
				return err
			}

			value.Map.ValueContract = contract.Element.ClonePointer()
		}

		return nil

	case BindingContractUnionKind:
		for index := range contract.Options {
			option := contract.Options[index].Clone()
			if err := option.CheckAndLearn(name, value); err == nil {
				contract.Options[index] = option
				return nil
			}
		}

		return fmt.Errorf(
			"binding %q expects %s value, got %s value",
			name,
			contract.ExpectedValueDescription(),
			valueKindName(value.Kind),
		)

	default:
		return fmt.Errorf("binding %q has unknown contract kind %d", name, contract.Kind)
	}
}

func checkArrayElementContract(name string, array *Array, elementContract *BindingContract) error {
	return array.ForEach(func(index *Number, value Value) error {
		if err := elementContract.CheckAndLearn(fmt.Sprintf("%s[%s]", name, index.Format(DefaultDecimalPlacesToDisplay)), value); err != nil {
			return err
		}

		return nil
	})
}

func checkMapValueContract(name string, m *Map, valueContract *BindingContract) error {
	return m.ForEach(func(key string, binding Binding) error {
		if err := valueContract.CheckAndLearn(fmt.Sprintf("%s[%q]", name, key), binding.Value); err != nil {
			return err
		}

		return nil
	})
}

func (contract BindingContract) ExpectedValueDescription() string {
	switch contract.Kind {
	case BindingContractAnyKind:
		return "any"
	case BindingContractSameKind:
		if contract.HasKindValue {
			return valueKindName(contract.KindValue)
		}
		return "same-kind"
	case BindingContractExactKind:
		return valueKindName(contract.KindValue)
	case BindingContractArrayKind:
		if contract.Element == nil {
			return "array"
		}
		return "array of " + contract.Element.ExpectedValueDescription()
	case BindingContractMapKind:
		if contract.Element == nil {
			return "map"
		}
		return "map of " + contract.Element.ExpectedValueDescription()
	case BindingContractUnionKind:
		parts := make([]string, 0, len(contract.Options))
		for _, option := range contract.Options {
			parts = append(parts, option.ExpectedValueDescription())
		}
		return joinContractDescriptions(parts)
	default:
		return "unknown"
	}
}

func joinContractDescriptions(parts []string) string {
	switch len(parts) {
	case 0:
		return "unknown"
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " or " + parts[1]
	default:
		return fmt.Sprintf("%s or %s", joinWithComma(parts[:len(parts)-1]), parts[len(parts)-1])
	}
}

func joinWithComma(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for _, part := range parts[1:] {
		result += ", " + part
	}

	return result
}

func valueKindName(kind ValueKind) string {
	switch kind {
	case ValueVoid:
		return "void"
	case ValueNumber:
		return "number"
	case ValueBool:
		return "bool"
	case ValueString:
		return "string"
	case ValueMap:
		return "map"
	case ValueArray:
		return "array"
	case ValueFunction:
		return "function"
	case ValueBuiltinFunction:
		return "builtin function"
	default:
		return "unknown"
	}
}
