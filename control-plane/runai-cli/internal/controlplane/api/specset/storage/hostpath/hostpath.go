package hostpath

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func NewList(list []string) ([]openapi.HostPathInstance, error) {
	var objList = make([]openapi.HostPathInstance, 0)
	for _, host := range list {
		obj, err := New(host)
		if err != nil {
			return nil, err
		}
		objList = append(objList, *obj)
	}
	return objList, nil
}

func New(att string) (*openapi.HostPathInstance, error) {
	hp := NewParser(att)
	err := hp.Parse()
	if err != nil {
		return nil, err
	}

	obj, err := create(hp)
	if err != nil {
		return nil, err
	}
	if hp.Get(ReadWrite) != nil {
		obj.SetReadOnly(false)
	} else {
		obj.SetReadOnly(true)
	}

	if hp.Get(MountPropagation) != nil {
		mp, err := openapi.NewHostPathMountPropagationFromValue(*hp.Get(MountPropagation))
		if err != nil {
			return nil, err
		}
		obj.SetMountPropagation(*mp)
	}
	return obj, nil
}

func create(hp *parser.AttributeMapper) (*openapi.HostPathInstance, error) {
	err := parser.CheckMandatoryFields(hp, Path, Mount)
	if err != nil {
		return nil, err
	}

	path := hp.Get(Path)
	mount := hp.Get(Mount)
	return &openapi.HostPathInstance{
		Path:      *openapi.NewNullableString(path),
		MountPath: *openapi.NewNullableString(mount),
	}, nil
}
