package main

import (
	"fmt"
	"strconv"
	"strings"
)

type BindingContractKind int

const (
	BindingContractAnyKind BindingContractKind = iota
	BindingContractSameKind
	BindingContractExactKind
	BindingContractArrayKind
	BindingContractMapKind
	BindingContractFunctionKind
	BindingContractUnionKind
	BindingContractAliasKind
)

type BindingContract struct {
	Kind               BindingContractKind
	KindValue          ValueKind
	HasKindValue       bool
	Element            *BindingContract
	Options            []BindingContract
	AliasName          string
	IsStructuredMap    bool
	MapFields          []BindingContractMapField
	MapWildcard        *BindingContract
	FunctionParameters []BindingContract
	FunctionReturn     *BindingContract
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

	if contract.FunctionParameters != nil {
		cloned.FunctionParameters = make([]BindingContract, len(contract.FunctionParameters))
		for index, parameter := range contract.FunctionParameters {
			cloned.FunctionParameters[index] = parameter.Clone()
		}
	}

	if contract.FunctionReturn != nil {
		functionReturn := contract.FunctionReturn.Clone()
		cloned.FunctionReturn = &functionReturn
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

	case BindingContractFunctionKind:
		if value.Kind != ValueFunction || value.Function == nil {
			return fmt.Errorf(
				"binding %q expects function value, got %s value",
				name,
				valueKindName(value.Kind),
			)
		}

		expectedCount := len(contract.FunctionParameters)
		if !functionCanAcceptArgumentCount(value.Function, expectedCount) {
			return fmt.Errorf(
				"binding %q expects function accepting %d argument(s)",
				name,
				expectedCount,
			)
		}

		if err := checkFunctionImplementationCompatibility(name, value.Function, contract); err != nil {
			return err
		}

		value.Function.SignatureContract = contract.ClonePointer()
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

func checkFunctionImplementationCompatibility(name string, fn *Function, signature *BindingContract) error {
	if fn == nil || signature == nil || signature.Kind != BindingContractFunctionKind {
		return nil
	}

	for index, promised := range signature.FunctionParameters {
		implementation, ok := functionImplementationParameterContract(fn, index)
		if !ok {
			return fmt.Errorf(
				"binding %q function parameter %d contract cannot be matched with signature parameter contract %s",
				name,
				index+1,
				promised.SourceString(),
			)
		}

		if !bindingContractAcceptsAll(implementation, promised) {
			return fmt.Errorf(
				"binding %q function parameter %d contract %s is not compatible with signature parameter contract %s",
				name,
				index+1,
				implementation.SourceString(),
				promised.SourceString(),
			)
		}
	}

	if fn.ReturnContract != nil && signature.FunctionReturn != nil &&
		!bindingContractAcceptsAll(*signature.FunctionReturn, *fn.ReturnContract) {
		return fmt.Errorf(
			"binding %q function return contract %s is not compatible with signature return contract %s",
			name,
			fn.ReturnContract.SourceString(),
			signature.FunctionReturn.SourceString(),
		)
	}

	return nil
}

func functionImplementationParameterContract(fn *Function, index int) (BindingContract, bool) {
	if fn.BytecodeFunction != nil {
		return bytecodeImplementationParameterContract(*fn.BytecodeFunction, index)
	}

	if index < len(fn.Parameters) {
		return fn.Parameters[index].Contract, true
	}

	if fn.VariadicParameter == nil {
		return BindingContract{}, false
	}

	return variadicElementImplementationContract(fn.VariadicParameter.Contract)
}

func bytecodeImplementationParameterContract(function BytecodeFunction, index int) (BindingContract, bool) {
	if index < len(function.Parameters) {
		return function.Parameters[index].Contract, true
	}

	if function.VariadicParameter == nil {
		return BindingContract{}, false
	}

	return variadicElementImplementationContract(function.VariadicParameter.Contract)
}

func variadicElementImplementationContract(contract BindingContract) (BindingContract, bool) {
	switch contract.Kind {
	case BindingContractAnyKind:
		return BindingContract{Kind: BindingContractAnyKind}, true
	case BindingContractArrayKind:
		if contract.Element == nil {
			return BindingContract{Kind: BindingContractAnyKind}, true
		}
		return contract.Element.Clone(), true
	case BindingContractSameKind:
		if !contract.HasKindValue {
			return BindingContract{Kind: BindingContractAnyKind}, true
		}
		if contract.KindValue == ValueArray {
			return BindingContract{Kind: BindingContractAnyKind}, true
		}
		return BindingContract{}, false
	default:
		return BindingContract{}, false
	}
}

func bindingContractAcceptsAll(accepting BindingContract, offered BindingContract) bool {
	if accepting.Kind == BindingContractAnyKind {
		return true
	}

	if offered.Kind == BindingContractUnionKind {
		for _, option := range offered.Options {
			if !bindingContractAcceptsAll(accepting, option) {
				return false
			}
		}
		return true
	}

	if accepting.Kind == BindingContractUnionKind {
		for _, option := range accepting.Options {
			if bindingContractAcceptsAll(option, offered) {
				return true
			}
		}
		return false
	}

	if offered.Kind == BindingContractAnyKind {
		return accepting.Kind == BindingContractAnyKind
	}

	if (accepting.Kind == BindingContractExactKind || accepting.Kind == BindingContractSameKind) &&
		acceptsSingleKind(accepting, offered) {
		return true
	}

	switch accepting.Kind {
	case BindingContractArrayKind:
		return arrayContractAcceptsAll(accepting, offered)
	case BindingContractMapKind:
		return mapContractAcceptsAll(accepting, offered)
	case BindingContractFunctionKind:
		return functionContractAcceptsAll(accepting, offered)
	default:
		return accepting.SourceString() == offered.SourceString()
	}
}

func acceptsSingleKind(accepting BindingContract, offered BindingContract) bool {
	acceptingKind, acceptingSingle := contractSingleValueKind(accepting)
	offeredKind, offeredSingle := contractSingleValueKind(offered)
	if !offeredSingle {
		return false
	}

	if accepting.Kind == BindingContractSameKind && !accepting.HasKindValue {
		return true
	}

	return acceptingSingle && acceptingKind == offeredKind
}

func contractSingleValueKind(contract BindingContract) (ValueKind, bool) {
	switch contract.Kind {
	case BindingContractExactKind:
		return contract.KindValue, true
	case BindingContractSameKind:
		if contract.HasKindValue {
			return contract.KindValue, true
		}
		return ValueVoid, false
	case BindingContractArrayKind:
		return ValueArray, true
	case BindingContractMapKind:
		return ValueMap, true
	case BindingContractFunctionKind:
		return ValueFunction, true
	case BindingContractUnionKind:
		if len(contract.Options) == 0 {
			return ValueVoid, false
		}

		kind, ok := contractSingleValueKind(contract.Options[0])
		if !ok {
			return ValueVoid, false
		}

		for _, option := range contract.Options[1:] {
			optionKind, ok := contractSingleValueKind(option)
			if !ok || optionKind != kind {
				return ValueVoid, false
			}
		}

		return kind, true
	default:
		return ValueVoid, false
	}
}

func arrayContractAcceptsAll(accepting BindingContract, offered BindingContract) bool {
	if offered.Kind != BindingContractArrayKind {
		return false
	}

	if accepting.Element == nil {
		return true
	}

	if offered.Element == nil {
		return false
	}

	return bindingContractAcceptsAll(*accepting.Element, *offered.Element)
}

func mapContractAcceptsAll(accepting BindingContract, offered BindingContract) bool {
	if offered.Kind != BindingContractMapKind {
		return false
	}

	if !accepting.IsStructuredMap && accepting.Element == nil {
		return true
	}

	if accepting.IsStructuredMap || offered.IsStructuredMap {
		return accepting.SourceString() == offered.SourceString()
	}

	if accepting.Element == nil {
		return true
	}

	if offered.Element == nil {
		return false
	}

	return bindingContractAcceptsAll(*accepting.Element, *offered.Element)
}

func functionContractAcceptsAll(accepting BindingContract, offered BindingContract) bool {
	if offered.Kind != BindingContractFunctionKind {
		return false
	}

	return accepting.SourceString() == offered.SourceString()
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
	case BindingContractFunctionKind:
		return "function matching " + contract.functionSignatureDescription()
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

	case BindingContractArrayKind, BindingContractMapKind, BindingContractFunctionKind:
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

		if contract.FunctionParameters != nil {
			resolved.FunctionParameters = make([]BindingContract, len(contract.FunctionParameters))
			for index := range contract.FunctionParameters {
				parameter, err := contract.FunctionParameters[index].resolveAliases(lookup, seen)
				if err != nil {
					return BindingContract{}, err
				}
				resolved.FunctionParameters[index] = parameter
			}
		}

		if contract.FunctionReturn != nil {
			functionReturn, err := contract.FunctionReturn.resolveAliases(lookup, seen)
			if err != nil {
				return BindingContract{}, err
			}
			resolved.FunctionReturn = &functionReturn
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
	case BindingContractArrayKind, BindingContractMapKind, BindingContractFunctionKind:
		if contract.Element != nil && contract.Element.HasAlias() {
			return true
		}
		for _, field := range contract.MapFields {
			if field.Contract.HasAlias() {
				return true
			}
		}
		if contract.MapWildcard != nil && contract.MapWildcard.HasAlias() {
			return true
		}
		for _, parameter := range contract.FunctionParameters {
			if parameter.HasAlias() {
				return true
			}
		}
		return contract.FunctionReturn != nil && contract.FunctionReturn.HasAlias()
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
	case BindingContractArrayKind, BindingContractMapKind, BindingContractFunctionKind:
		if contract.Element != nil {
			contract.Element.collectAliasNames(names, seen)
		}
		for _, field := range contract.MapFields {
			field.Contract.collectAliasNames(names, seen)
		}
		if contract.MapWildcard != nil {
			contract.MapWildcard.collectAliasNames(names, seen)
		}
		for _, parameter := range contract.FunctionParameters {
			parameter.collectAliasNames(names, seen)
		}
		if contract.FunctionReturn != nil {
			contract.FunctionReturn.collectAliasNames(names, seen)
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
	case BindingContractFunctionKind:
		return "STD.TYPE.FUNCTION<" + contract.functionSignatureSourceString() + ">"
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

func checkFunctionSignatureArguments(contract *BindingContract, args []Value) error {
	if contract == nil || contract.Kind != BindingContractFunctionKind {
		return nil
	}

	expectedCount := len(contract.FunctionParameters)
	if len(args) != expectedCount {
		return fmt.Errorf("function contract expected %d argument(s), got %d", expectedCount, len(args))
	}

	for index := range contract.FunctionParameters {
		if err := contract.FunctionParameters[index].CheckAndLearn(
			fmt.Sprintf("function argument %d", index+1),
			args[index],
		); err != nil {
			return err
		}
	}

	return nil
}

func checkFunctionSignatureReturn(contract *BindingContract, value Value) (Value, error) {
	if contract == nil || contract.Kind != BindingContractFunctionKind || contract.FunctionReturn == nil {
		return value, nil
	}

	if err := contract.FunctionReturn.CheckAndLearn("function return", value); err != nil {
		return Value{}, err
	}

	return value, nil
}

func checkFunctionReturnContracts(
	returnContract *BindingContract,
	signatureContract *BindingContract,
	value Value,
) (Value, error) {
	if returnContract != nil {
		if err := returnContract.CheckAndLearn("function return", value); err != nil {
			return Value{}, err
		}
	}

	return checkFunctionSignatureReturn(signatureContract, value)
}

func (contract BindingContract) functionSignatureDescription() string {
	parameterDescriptions := make([]string, 0, len(contract.FunctionParameters))
	for _, parameter := range contract.FunctionParameters {
		parameterDescriptions = append(parameterDescriptions, parameter.ExpectedValueDescription())
	}

	returnDescription := "void"
	if contract.FunctionReturn != nil {
		returnDescription = contract.FunctionReturn.ExpectedValueDescription()
	}

	return "(" + joinContractMapSourceParts(parameterDescriptions) + ") returning " + returnDescription
}

func (contract BindingContract) functionSignatureSourceString() string {
	parameterSources := make([]string, 0, len(contract.FunctionParameters))
	for _, parameter := range contract.FunctionParameters {
		parameterSources = append(parameterSources, parameter.SourceString())
	}

	returnSource := "STD.TYPE.VOID"
	if contract.FunctionReturn != nil {
		returnSource = contract.FunctionReturn.SourceString()
	}

	return "(" + joinContractMapSourceParts(parameterSources) + "): " + returnSource
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

	var result strings.Builder

	result.WriteString(parts[0])

	for _, part := range parts[1:] {
		result.WriteByte('|')
		result.WriteString(part)
	}

	return result.String()
}
