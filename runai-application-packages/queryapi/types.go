package queryapi

type APIFieldsMapping map[string]FieldMapping

func NewFieldMapping(column string, validators ...Validator) FieldMapping {
	validators = append(validators, UserInputValidator{})
	return FieldMapping{
		Column:     column,
		Validators: validators,
	}
}

type FieldMapping struct {
	Column     string
	Validators []Validator
}

type Filter struct {
	Field       string
	Operator    string
	ValueFormat string
	Values      []string
}

type Sort struct {
	Field     string
	Direction string
}

type operatorInfo struct {
	Operator    string
	ValueFormat string
}
