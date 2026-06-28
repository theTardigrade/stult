package main

import (
	"fmt"
	"strconv"
)

type BindingContractKind int

const (
	BindingContractAnyKind BindingContractKind = iota
	BindingContractSameKind
	BindingContractExactKind
	BindingContractArrayKind
	BindingContractMapKind
	BindingContractUnionKind
	BindingContractAliasKind
)

type BindingContract struct {
	Kind            BindingContractKind
	KindValue       ValueKind
	HasKindValue    bool
	Element         *BindingContract
	Options         []BindingContract
	AliasName       string
	IsStructuredMap bool
	MapFields       []BindingContractMapField
	MapWildcard     *BindingContract
}

type BindingContractMapField struct {
	Key        string
	Contract   BindingContract
	IsOptional bool
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

	if contract.MapFields != nil {
		cloned.MapFields = make([]BindingContractMapField, len(contract.MapFields))
		for index, field := range contract.MapFields {
			cloned.MapFields[index] = BindingContractMapField{
				Key:        field.Key,
				Contract:   field.Contract.Clone(),
				IsOptional: field.IsOptional,
			}
		}
	}

	if contract.MapWildcard != nil {
		wildcard := contract.MapWildcard.Clone()
		cloned.MapWildcard = &wildcard
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

		if contract.IsStructuredMap {
			if err := checkStructuredMapContract(name, value.Map, contract); err != nil {
				return err
			}

			value.Map.StructuredContract = contract.ClonePointer()
			return nil
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

	case BindingContractAliasKind:
		return fmt.Errorf("binding %q has unresolved contract alias %q", name, contract.AliasName)

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

func checkStructuredMapContract(name string, m *Map, contract *BindingContract) error {
	fieldIndexes := map[string]int{}
	for index, field := range contract.MapFields {
		if _, exists := fieldIndexes[field.Key]; exists {
			return fmt.Errorf("binding %q has duplicate map contract key %q", name, field.Key)
		}
		fieldIndexes[field.Key] = index

		binding, exists := m.Get(field.Key)
		if !exists {
			if field.IsOptional {
				continue
			}

			return fmt.Errorf("binding %q missing required map key %q", name, field.Key)
		}

		if err := contract.MapFields[index].Contract.CheckAndLearn(fmt.Sprintf("%s[%q]", name, field.Key), binding.Value); err != nil {
			return err
		}
	}

	return m.ForEach(func(key string, binding Binding) error {
		if _, exists := fieldIndexes[key]; exists {
			return nil
		}

		if contract.MapWildcard != nil {
			return contract.MapWildcard.CheckAndLearn(fmt.Sprintf("%s[%q]", name, key), binding.Value)
		}

		return fmt.Errorf("binding %q does not allow map key %q", name, key)
	})
}

func (contract *BindingContract) CheckStructuredMapEntryAndLearn(name string, key string, value Value) error {
	if contract == nil || contract.Kind != BindingContractMapKind || !contract.IsStructuredMap {
		return nil
	}

	for index := range contract.MapFields {
		field := contract.MapFields[index]
		if field.Key != key {
			continue
		}

		return contract.MapFields[index].Contract.CheckAndLearn(fmt.Sprintf("%s[%q]", name, key), value)
	}

	if contract.MapWildcard != nil {
		return contract.MapWildcard.CheckAndLearn(fmt.Sprintf("%s[%q]", name, key), value)
	}

	return fmt.Errorf("%s does not allow map key %q", name, key)
}

func (contract *BindingContract) HasRequiredStructuredMapFields() bool {
	if contract == nil || contract.Kind != BindingContractMapKind || !contract.IsStructuredMap {
		return false
	}

	for _, field := range contract.MapFields {
		if !field.IsOptional {
			return true
		}
	}

	return false
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
		if contract.IsStructuredMap {
			return "map matching " + contract.structuredMapDescription()
		}
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
	case BindingContractAliasKind:
		return contract.AliasName
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
	case ValueContract:
		return "contract"
	default:
		return "unknown"
	}
}

type BindingContractValueLookup func(name string) (Value, bool)

func NewContractValue(contract BindingContract) Value {
	cloned := contract.Clone()
	return Value{Kind: ValueContract, Contract: &cloned}
}

func (contract BindingContract) ResolveAliases(lookup BindingContractValueLookup) (BindingContract, error) {
	return contract.resolveAliases(lookup, map[string]bool{})
}

func (contract BindingContract) resolveAliases(
	lookup BindingContractValueLookup,
	seen map[string]bool,
) (BindingContract, error) {
	switch contract.Kind {
	case BindingContractAliasKind:
		if contract.AliasName == "" {
			return BindingContract{}, fmt.Errorf("empty contract alias")
		}

		if seen[contract.AliasName] {
			return BindingContract{}, fmt.Errorf("cyclic contract alias %q", contract.AliasName)
		}

		value, ok := lookup(contract.AliasName)
		if !ok {
			return BindingContract{}, fmt.Errorf("undefined contract alias %q", contract.AliasName)
		}

		value = resolveSpecializedValue(value)
		if value.Kind != ValueContract || value.Contract == nil {
			return BindingContract{}, fmt.Errorf("contract alias %q must be a contract value, got %s value", contract.AliasName, valueKindName(value.Kind))
		}

		seen[contract.AliasName] = true
		resolved, err := value.Contract.resolveAliases(lookup, seen)
		delete(seen, contract.AliasName)
		return resolved, err

	case BindingContractArrayKind, BindingContractMapKind:
		resolved := contract.Clone()
		if contract.Element != nil {
			element, err := contract.Element.resolveAliases(lookup, seen)
			if err != nil {
				return BindingContract{}, err
			}
			resolved.Element = &element
		}

		if contract.MapFields != nil {
			resolved.MapFields = make([]BindingContractMapField, len(contract.MapFields))
			for index, field := range contract.MapFields {
				fieldContract, err := field.Contract.resolveAliases(lookup, seen)
				if err != nil {
					return BindingContract{}, err
				}
				resolved.MapFields[index] = BindingContractMapField{
					Key:        field.Key,
					Contract:   fieldContract,
					IsOptional: field.IsOptional,
				}
			}
		}

		if contract.MapWildcard != nil {
			wildcard, err := contract.MapWildcard.resolveAliases(lookup, seen)
			if err != nil {
				return BindingContract{}, err
			}
			resolved.MapWildcard = &wildcard
		}

		return resolved, nil

	case BindingContractUnionKind:
		resolved := contract.Clone()
		resolved.Options = make([]BindingContract, len(contract.Options))
		for index := range contract.Options {
			option, err := contract.Options[index].resolveAliases(lookup, seen)
			if err != nil {
				return BindingContract{}, err
			}
			resolved.Options[index] = option
		}
		return resolved, nil

	default:
		return contract.Clone(), nil
	}
}

func (contract BindingContract) HasAlias() bool {
	switch contract.Kind {
	case BindingContractAliasKind:
		return true
	case BindingContractArrayKind, BindingContractMapKind:
		if contract.Element != nil && contract.Element.HasAlias() {
			return true
		}
		for _, field := range contract.MapFields {
			if field.Contract.HasAlias() {
				return true
			}
		}
		return contract.MapWildcard != nil && contract.MapWildcard.HasAlias()
	case BindingContractUnionKind:
		for _, option := range contract.Options {
			if option.HasAlias() {
				return true
			}
		}
	}
	return false
}

func (contract BindingContract) AliasNames() []string {
	names := []string{}
	seen := map[string]bool{}
	contract.collectAliasNames(&names, seen)
	return names
}

func (contract BindingContract) collectAliasNames(names *[]string, seen map[string]bool) {
	switch contract.Kind {
	case BindingContractAliasKind:
		if contract.AliasName != "" && !seen[contract.AliasName] {
			seen[contract.AliasName] = true
			*names = append(*names, contract.AliasName)
		}
	case BindingContractArrayKind, BindingContractMapKind:
		if contract.Element != nil {
			contract.Element.collectAliasNames(names, seen)
		}
		for _, field := range contract.MapFields {
			field.Contract.collectAliasNames(names, seen)
		}
		if contract.MapWildcard != nil {
			contract.MapWildcard.collectAliasNames(names, seen)
		}
	case BindingContractUnionKind:
		for _, option := range contract.Options {
			option.collectAliasNames(names, seen)
		}
	}
}

func (contract BindingContract) SourceString() string {
	switch contract.Kind {
	case BindingContractAnyKind:
		return "*"
	case BindingContractSameKind:
		return "."
	case BindingContractExactKind:
		switch contract.KindValue {
		case ValueVoid:
			return "STD.TYPE.VOID"
		case ValueNumber:
			return "STD.TYPE.NUMBER"
		case ValueBool:
			return "STD.TYPE.BOOL"
		case ValueString:
			return "STD.TYPE.STRING"
		case ValueFunction:
			return "STD.TYPE.FUNCTION"
		case ValueBuiltinFunction:
			return "STD.TYPE.BUILTIN_FUNCTION"
		case ValueContract:
			return "STD.TYPE.CONTRACT"
		default:
			return valueKindName(contract.KindValue)
		}
	case BindingContractArrayKind:
		if contract.Element == nil {
			return "STD.TYPE.ARRAY"
		}
		return "STD.TYPE.ARRAY<" + contract.Element.SourceString() + ">"
	case BindingContractMapKind:
		if contract.IsStructuredMap {
			return "STD.TYPE.MAP<" + contract.structuredMapSourceString() + ">"
		}
		if contract.Element == nil {
			return "STD.TYPE.MAP"
		}
		return "STD.TYPE.MAP<" + contract.Element.SourceString() + ">"
	case BindingContractUnionKind:
		parts := make([]string, 0, len(contract.Options))
		for _, option := range contract.Options {
			parts = append(parts, option.SourceString())
		}
		return joinContractSourceParts(parts)
	case BindingContractAliasKind:
		return contract.AliasName
	default:
		return "unknown"
	}
}

func (contract BindingContract) structuredMapDescription() string {
	parts := []string{}
	for _, field := range contract.MapFields {
		marker := ""
		if field.IsOptional {
			marker = " optional"
		}
		parts = append(parts, fmt.Sprintf("%q%s as %s", field.Key, marker, field.Contract.ExpectedValueDescription()))
	}

	if contract.MapWildcard != nil {
		parts = append(parts, "other keys as "+contract.MapWildcard.ExpectedValueDescription())
	}

	if len(parts) == 0 {
		return "{}"
	}

	return "{" + joinWithComma(parts) + "}"
}

func (contract BindingContract) structuredMapSourceString() string {
	parts := []string{}
	for _, field := range contract.MapFields {
		key := strconv.Quote(field.Key)
		if isIdentifierString(field.Key) && field.Key != "_" {
			key = "." + field.Key
		}
		if field.IsOptional {
			key += "?"
		}
		parts = append(parts, key+": "+field.Contract.SourceString())
	}

	if contract.MapWildcard != nil {
		parts = append(parts, "_: "+contract.MapWildcard.SourceString())
	}

	return "{" + joinContractMapSourceParts(parts) + "}"
}

func joinContractMapSourceParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += ", " + part
	}
	return result
}

func isIdentifierString(text string) bool {
	if text == "" {
		return false
	}

	for index, ch := range text {
		if index == 0 {
			if !isIdentStart(ch) {
				return false
			}
			continue
		}

		if !isIdentPart(ch) {
			return false
		}
	}

	return true
}

func joinContractSourceParts(parts []string) string {
	if len(parts) == 0 {
		return "unknown"
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += "|" + part
	}
	return result
}
