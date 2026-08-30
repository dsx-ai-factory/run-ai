package cmv

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Name        = "name"
	Path        = "path"
	SubPath     = "subpath"
	DefaultMode = "default-mode"

	ParserName = "ConfigMapVolumeParser"
)

var (
	ConfigMapVolume = []string{Name, Path, SubPath, DefaultMode}
)

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](ConfigMapVolume...)))
}
