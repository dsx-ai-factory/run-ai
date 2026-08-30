package json_patch

import "encoding/json"

type operationType string

const (
	operationAdd     operationType = "add"
	operationRemove  operationType = "remove"
	operationReplace operationType = "replace"
	operationCopy    operationType = "copy"
	operationMove    operationType = "move"
	operationTest    operationType = "test"
)

type patch struct {
	operations []*operation
}

func NewPatch() *patch {
	return &patch{}
}

func (p *patch) AddOperations(operations ...*operation) {
	p.operations = append(p.operations, operations...)
}

type operation struct {
	Operation operationType `json:"op"`

	From string `json:"from,omitempty"`

	Path string `json:"path"`

	Value interface{} `json:"value,omitempty"`
}

// NewReplaceOperation creates a replacing json patch
func NewReplaceOperation(path string, value interface{}) *operation {
	return &operation{
		Operation: operationReplace,
		Path:      path,
		Value:     value,
	}
}

// Bytes returns the patch raw data
func (p *patch) Bytes() ([]byte, error) {
	bytes, err := json.Marshal(p.operations)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
