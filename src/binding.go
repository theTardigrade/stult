package main

import "fmt"

type BindingContractKind int

const (
	BindingContractAnyKind BindingContractKind = iota
	BindingContractSameKind
)

type BindingContract struct {
	Kind      BindingContractKind
	KindValue ValueKind
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
		Kind:      BindingContractSameKind,
		KindValue: value.Kind,
	}
}

func BindingContractFromToken(token Token, initialValue Value) BindingContract {
	if token.Type == TokenContractSameKind {
		return NewSameKindBindingContract(initialValue)
	}

	return BindingContract{}
}

func (contract BindingContract) IsAny() bool {
	return contract.Kind == BindingContractAnyKind
}

func (contract BindingContract) Check(name string, value Value) error {
	value = resolveSpecializedValue(value)

	switch contract.Kind {
	case BindingContractAnyKind:
		return nil

	case BindingContractSameKind:
		if value.Kind == contract.KindValue {
			return nil
		}

		return fmt.Errorf(
			"binding %q expects %s value, got %s value",
			name,
			valueKindName(contract.KindValue),
			valueKindName(value.Kind),
		)

	default:
		return fmt.Errorf("binding %q has unknown contract kind %d", name, contract.Kind)
	}
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
