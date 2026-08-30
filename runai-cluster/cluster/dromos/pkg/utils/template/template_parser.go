package template

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a runtime.RawExtension template into an unstructured.Unstructured object
// with the GVK properly set.
func (tp *Parser) Parse(template runtime.RawExtension) (*unstructured.Unstructured, error) {
	if len(template.Raw) == 0 {
		return nil, errors.New("template is empty")
	}

	var obj unstructured.Unstructured
	if err := json.Unmarshal(template.Raw, &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	if obj.GetAPIVersion() == "" || obj.GetKind() == "" {
		return nil, errors.New("template missing required apiVersion or kind fields")
	}

	return &obj, nil
}
