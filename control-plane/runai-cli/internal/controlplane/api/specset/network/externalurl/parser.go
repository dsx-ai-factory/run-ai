package externalurl

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Container  = "container"
	Url        = "url"
	AuthUsers  = "authusers"
	AuthGroups = "authgroups"

	ParserName = "ExternalUrlParser"
)

var (
	ExternalUrl = []string{Container, Url, AuthUsers, AuthGroups}
)

// container=9443,url=https://external.runai.com,authusers=user1,authgroups=group1

func NewParser(port string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(port, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](ExternalUrl...)))
}
