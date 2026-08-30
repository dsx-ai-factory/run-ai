package field_types

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"strconv"
)

type Field interface {
	IsValueInitialized() bool
	String() string
}

// +k8s:deepcopy-gen=true
type IntField struct {
	Value *int      `json:"value,omitempty"`
	Rules *IntRules `json:"rules,omitempty"`
}

// +k8s:deepcopy-gen=true
type StringField struct {
	Value *string      `json:"value,omitempty"`
	Rules *StringRules `json:"rules,omitempty"`
}

// +k8s:deepcopy-gen=true
type QuantityField struct {
	//+kubebuilder:validation:Pattern:=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	Value *string        `json:"value,omitempty"`
	Rules *QuantityRules `json:"rules,omitempty"`
}

func (field *QuantityField) Deref() string {
	return field.DerefOrDefault("")
}

func (field *QuantityField) DerefOrDefault(def string) string {
	if field == nil || field.Value == nil {
		return def
	} else {
		return *field.Value
	}
}

func (field *QuantityField) DerefQuantity(defaultValue int64) (resource.Quantity, error) {
	if field == nil || field.Value == nil {
		return *resource.NewQuantity(defaultValue, resource.DecimalSI), nil
	}
	return resource.ParseQuantity(*field.Value)
}

func (field *QuantityField) IsValueInitialized() bool {
	return field != nil && field.Value != nil
}

func (field *QuantityField) String() string {
	return field.Deref()
}

// +k8s:deepcopy-gen=true
type BoolField struct {
	Value *bool      `json:"value,omitempty"`
	Rules *BoolRules `json:"rules,omitempty"`
}

// +k8s:deepcopy-gen=true
type NumberField struct {
	//+kubebuilder:validation:Format:=^[+-]?([0-9]+\.?[0-9]*|\.[0-9]+)$
	Value *string      `json:"value,omitempty"`
	Rules *NumberRules `json:"rules,omitempty"`
}

func (field *StringField) Deref() string {
	return field.DerefOrDefault("")
}

func (field *StringField) DerefOrDefault(def string) string {
	if field == nil || field.Value == nil {
		return def
	} else {
		return *field.Value
	}
}

func (field *StringField) IsValueInitialized() bool {
	return field != nil && field.Value != nil
}

func (field *StringField) String() string {
	return field.Deref()
}

func (field *BoolField) Deref() bool {
	if field == nil || field.Value == nil {
		return false
	} else {
		return *field.Value
	}
}

func (field *BoolField) IsValueInitialized() bool {
	return field != nil && field.Value != nil
}

func (field *BoolField) String() string {
	return strconv.FormatBool(field.Deref())
}

func (field *IntField) Deref() int {
	if field == nil || field.Value == nil {
		return 0
	} else {
		return *field.Value
	}
}

func (field *IntField) Deref32() int32 {
	return int32(field.Deref())
}

func (field *IntField) Deref32default(defaultVal int32) int32 {
	if field == nil || field.Value == nil {
		return defaultVal
	} else {
		return field.Deref32()
	}
}

func (field *IntField) IsValueInitialized() bool {
	return field != nil && field.Value != nil
}

func (field *IntField) String() string {
	return strconv.Itoa(field.Deref())
}

func (field *NumberField) DerefStr() string {
	if field == nil || field.Value == nil {
		return ""
	} else {
		return *field.Value
	}
}

func (field *NumberField) Deref() float64 {
	if field == nil || field.Value == nil {
		return 0.0
	} else {
		// I do not check error here, because it is already been validated
		result, _ := strconv.ParseFloat(*field.Value, 32)
		return result
	}
}

func (field *NumberField) IsValueInitialized() bool {
	return field != nil && field.Value != nil
}

func (field *NumberField) String() string {
	return field.DerefStr()
}

func NewStringField(value string) *StringField {
	return &StringField{
		Value: ptr.To(value),
	}
}

func NewOptionalStringField(value string, defaultValue string) *StringField {
	if value == defaultValue {
		return nil
	}
	return &StringField{
		Value: ptr.To(value),
	}
}

func NewQuantityField(value string) *QuantityField {
	return &QuantityField{
		Value: ptr.To(value),
	}
}

func NewOptionalQuantityField(value string, defaultValue string) *QuantityField {
	if value == defaultValue {
		return nil
	}
	return &QuantityField{
		Value: ptr.To(value),
	}
}

func NewOptionalNumberField(value *float64) *NumberField {
	if value == nil {
		return nil
	}
	return &NumberField{
		Value: ptr.To(fmt.Sprintf("%g", *value)),
	}
}

func NewOptionalIntField(value *int) *IntField {
	if value == nil {
		return nil
	}
	return &IntField{
		Value: value,
	}
}

func NewOptionalBoolField(value *bool) *BoolField {
	if value == nil {
		return nil
	}
	return &BoolField{
		Value: value,
	}
}
