package main

func stdExactContract(kind ValueKind) BindingContract {
	return BindingContract{
		Kind:      BindingContractExactKind,
		KindValue: kind,
	}
}

func stdArrayContract(element BindingContract) BindingContract {
	return BindingContract{
		Kind:    BindingContractArrayKind,
		Element: element.ClonePointer(),
	}
}

func stdUnionContract(options ...BindingContract) BindingContract {
	cloned := make([]BindingContract, len(options))
	for index := range options {
		cloned[index] = options[index].Clone()
	}

	return BindingContract{
		Kind:    BindingContractUnionKind,
		Options: cloned,
	}
}

func stdStructuredMapContract(fields ...BindingContractMapField) BindingContract {
	cloned := make([]BindingContractMapField, len(fields))
	for index, field := range fields {
		cloned[index] = BindingContractMapField{
			Key:        field.Key,
			Contract:   field.Contract.Clone(),
			IsOptional: field.IsOptional,
		}
	}

	return BindingContract{
		Kind:            BindingContractMapKind,
		IsStructuredMap: true,
		MapFields:       cloned,
	}
}

func stdRequiredMapField(key string, contract BindingContract) BindingContractMapField {
	return BindingContractMapField{Key: key, Contract: contract.Clone()}
}

func stdOptionalMapField(key string, contract BindingContract) BindingContractMapField {
	return BindingContractMapField{Key: key, Contract: contract.Clone(), IsOptional: true}
}
