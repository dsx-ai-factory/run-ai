package gitsync

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func NewGitSyncList(gitList []string) ([]openapi.GitInstance, error) {
	var gitObjList = make([]openapi.GitInstance, 0)
	for _, git := range gitList {
		gitSync, err := NewGit(git)
		if err != nil {
			return nil, err
		}
		gitObjList = append(gitObjList, *gitSync)
	}
	return gitObjList, nil
}

func NewGit(git string) (*openapi.GitInstance, error) {
	gp := NewParser(git)
	err := gp.Parse()
	if err != nil {
		return nil, err
	}

	gitSync, err := createGit(gp)
	if err != nil {
		return nil, err
	}
	secret := gp.Get(Secret)
	if secret != nil {
		gitSync.SetPasswordSecret(*secret)
	}
	rev := gp.Get(Revision)
	if rev != nil {
		gitSync.SetRevision(*rev)
	}
	return gitSync, nil
}

func createGit(gp *parser.AttributeMapper) (*openapi.GitInstance, error) {
	err := parser.CheckMandatoryFields(gp, Repository, Path)
	if err != nil {
		return nil, err
	}

	name := gp.Get(Name)
	repo := gp.Get(Repository)
	path := gp.Get(Path)
	return &openapi.GitInstance{
		Name:       *openapi.NewNullableString(name),
		Repository: *openapi.NewNullableString(repo),
		Path:       *openapi.NewNullableString(path),
	}, nil
}
