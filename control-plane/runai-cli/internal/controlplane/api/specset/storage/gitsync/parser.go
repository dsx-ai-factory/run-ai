package gitsync

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Name       = "name"
	Repository = "repository"
	Path       = "path"
	Secret     = "secret"
	Revision   = "rev"

	ParserName = "GitSyncParser"
)

var (
	GitSync = []string{Name, Repository, Path, Secret, Revision}
)

// name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](GitSync...)))
}
