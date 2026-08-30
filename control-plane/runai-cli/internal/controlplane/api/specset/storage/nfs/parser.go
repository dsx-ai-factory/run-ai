package nfs

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Path      = "path"
	Server    = "server"
	MountPath = "mountpath"
	ReadWrite = "readwrite"

	ParserName = "NfsParser"
)

var (
	Nfs = []string{Path, Server, MountPath, ReadWrite}
)

// path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](Nfs...)))
}
