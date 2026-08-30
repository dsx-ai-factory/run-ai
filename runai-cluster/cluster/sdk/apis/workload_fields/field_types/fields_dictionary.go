package field_types

import (
	"k8s.io/utils/ptr"
	"sort"
)

const DictionaryFieldDeleted = "-" // same as in kubectl (e.g. when deleting a label)

// +kubebuilder:object:generate:=true
type DictionaryField struct {
	Rules     *MapRules                          `json:"rules,omitempty"`
	ItemRules *StringRules                       `json:"itemRules,omitempty"`
	KeyRules  *StringRules                       `json:"keyRules,omitempty"`
	Value     *map[string]DictionaryItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type DictionaryItemAndRules struct {
	Rules   *MapItemRules `json:"rules,omitempty"`
	Deleted *bool         `json:"deleted,omitempty"`
	Value   *string       `json:"value,omitempty"`
}

func (self *DictionaryField) Deref() map[string]DictionaryItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]DictionaryItemAndRules{}
	}
}

func NewDictionaryField(value map[string]string) *DictionaryField {
	items := make(map[string]DictionaryItemAndRules)
	for key, value := range value {
		if value != DictionaryFieldDeleted {
			items[key] = DictionaryItemAndRules{
				Value: ptr.To(value),
			}
		} else {
			items[key] = DictionaryItemAndRules{
				Deleted: ptr.To(true),
			}
		}
	}
	return &DictionaryField{
		Value: &items,
	}
}

func (self *DictionaryField) Put(key, value string) {
	if self.Value == nil {
		self.Value = &map[string]DictionaryItemAndRules{}
	}
	if value != DictionaryFieldDeleted {
		(*self.Value)[key] = DictionaryItemAndRules{
			Value: ptr.To(value),
		}
	} else {
		(*self.Value)[key] = DictionaryItemAndRules{
			Deleted: ptr.To(true),
		}
	}
}

func SortedDictionaryMapKeys(unsorted map[string]DictionaryItemAndRules) []string {
	result := make([]string, 0, len(unsorted))
	for k := range unsorted {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
