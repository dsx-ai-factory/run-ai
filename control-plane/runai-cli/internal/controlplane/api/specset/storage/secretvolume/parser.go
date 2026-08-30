package secretvolume

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	MountPath   = "path"
	SecretName  = "name"
	DefaultMode = "default-mode"
	ParserName  = "SecretVolumeParser"
)

var (
	SecretVolume = []string{MountPath, SecretName, DefaultMode}
)

// path=MountPath,name=SecretName

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](SecretVolume...)))
}
