package datavolumes

import (
	"context"

	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

type SimpleMock struct {
	DataVolumes []openapi.DatavolumeNoSharedScopes
}

func NewSimpleMock() *SimpleMock {
	return &SimpleMock{
		DataVolumes: make([]openapi.DatavolumeNoSharedScopes, 0),
	}
}

func (smc *SimpleMock) GetDataVolumesNotPaginated(_ context.Context, _ openapi.DatavolumeRequestType, _ string,
	_ []string, _ string, _ string) ([]openapi.DatavolumeNoSharedScopes, error) {
	return smc.DataVolumes, nil
}

func (smc *SimpleMock) WithDataVolumes(dataVolumes []openapi.DatavolumeNoSharedScopes) *SimpleMock {
	smc.DataVolumes = dataVolumes
	return smc
}
