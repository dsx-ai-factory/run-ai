package nfs

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func NewNfsList(hostList []string) ([]openapi.NfsInstance, error) {
	var hostObjList = make([]openapi.NfsInstance, 0)
	for _, host := range hostList {
		gitSync, err := NewNfs(host)
		if err != nil {
			return nil, err
		}
		hostObjList = append(hostObjList, *gitSync)
	}
	return hostObjList, nil
}

func NewNfs(nfs string) (*openapi.NfsInstance, error) {
	np := NewParser(nfs)
	err := np.Parse()
	if err != nil {
		return nil, err
	}

	nfsObj, err := createNfs(np)
	if err != nil {
		return nil, err
	}
	if np.Get(ReadWrite) != nil {
		nfsObj.SetReadOnly(false)
	} else {
		nfsObj.SetReadOnly(true)
	}
	return nfsObj, nil
}

func createNfs(np *parser.AttributeMapper) (*openapi.NfsInstance, error) {
	err := parser.CheckMandatoryFields(np, Path, Server, MountPath)
	if err != nil {
		return nil, err
	}

	path := np.Get(Path)
	server := np.Get(Server)
	mountPath := np.Get(MountPath)
	return &openapi.NfsInstance{
		Path:      *openapi.NewNullableString(path),
		Server:    *openapi.NewNullableString(server),
		MountPath: *openapi.NewNullableString(mountPath),
	}, nil
}
