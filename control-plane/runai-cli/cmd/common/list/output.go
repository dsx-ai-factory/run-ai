package list

import (
	"errors"
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config/attributes"
	"go.uber.org/zap"
)

const (
	errMsgSoleRaw     = "raw flag must be combined with json or yaml flag"
	errMsgUnknownType = "unknown output type: %s, default will be used"
)

type GenerateOutput interface {
	GenerateJson() (string, error)
	GenerateYaml() (string, error)
	GenerateTable() (string, error)
}

type RawSetter interface {
	SetRaw(raw bool)
}
type ColumnsSetter interface {
	SetColumns(columns []string) error
}

type Output struct {
	ofJson              bool
	ofYaml              bool
	ofTable             bool
	ofTableNoHeaders    bool
	ofRaw               bool
	ofSetDefaultColumns bool
	ofColumns           []string
}

// OutputMapAligned represents a data row in JSON format with column titles as keys
type OutputMapAligned map[string]interface{}

func (o *Output) IsJson() bool {
	return o.ofJson
}

func (o *Output) IsYaml() bool {
	return o.ofYaml
}

func (o *Output) IsTable() bool {
	return o.ofTable
}

func (o *Output) IsRaw() bool {
	return o.ofRaw
}
func (o *Output) IsSetDefaultColumns() bool {
	return o.ofSetDefaultColumns
}

func (o *Output) GetColumns() []string {
	return o.ofColumns
}

func (o *Output) SetDefault(value string) error {
	logger := config.GetInstance().GetLogger()
	if o.ofJson || o.ofYaml || o.ofTable {
		logger.Debug("output type was set, not default will be used")
		return nil
	}

	switch value {
	case attributes.EmptyString:
		logger.Debug("output type not set, default will be used (table)")
	case JsonStr:
		o.ofJson = true
	case YamlStr:
		o.ofYaml = true
	case TableStr:
		o.ofTable = true
	default:
		return fmt.Errorf(errMsgUnknownType, value)
	}
	logger.Debug("output type set", zap.String("output", value))
	return nil
}

func (o *Output) Generate(generate GenerateOutput) (string, error) {
	// If the GenerateOutput implementation supports setting raw preference, set it
	// temp check until we implement raw flag in all the lists
	if rawSettable, ok := generate.(RawSetter); ok {
		rawSettable.SetRaw(o.ofRaw)
	}

	if columnsSettable, ok := generate.(ColumnsSetter); ok {
		err := columnsSettable.SetColumns(o.ofColumns)
		if err != nil {
			return "", err
		}

	}

	if o.IsJson() {
		return generate.GenerateJson()
	}

	if o.ofYaml {
		return generate.GenerateYaml()
	}

	generated, err := generate.GenerateTable()
	if err != nil {
		return "", err
	}

	if o.ofTableNoHeaders {
		return removeHeaders(generated), nil
	}
	return generated, nil
}

func removeHeaders(input string) string {
	lines := strings.Split(input, "\n")

	// Check if there are at least 3 lines before trimming
	if len(lines) < 3 {
		return ""
	}

	return strings.Join(lines[2:], "\n")
}

func (o *Output) RawFlagValidator() error {
	if o.IsRaw() && !o.IsJson() && !o.IsYaml() {
		return errors.New(errMsgSoleRaw)
	}
	return nil
}
