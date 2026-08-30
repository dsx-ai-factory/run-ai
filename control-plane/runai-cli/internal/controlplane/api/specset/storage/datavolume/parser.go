package datavolume

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Name      = "name"
	MountPath = "mountpath"

	ParserName = "DataVolumeParser"
)

var (
	DataVolumeAttributes = []string{Name, MountPath}
)

func NewParser(dv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(dv, parser.NewStrictMapParser(ParserName, mapset.NewSet[string](DataVolumeAttributes...)))
}
