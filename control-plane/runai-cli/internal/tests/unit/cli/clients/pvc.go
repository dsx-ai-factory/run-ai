package clients

import (
	"github.com/golang/mock/gomock"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/pvc"
	openapi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
)

type MockPvcResponse struct {
	Response *openapi.PVCListResponse
	Err      error
}

func MockPvc(mockCtrl *gomock.Controller, projPvcMap map[string]MockPvcResponse) {
	mockK8sObjTracker := cpmocks.NewMockPvcLister(mockCtrl)
	for key, value := range projPvcMap {
		mockK8sObjTracker.EXPECT().List(gomock.Any(), key).Return(value.Response, value.Err).AnyTimes()
	}
	pvc.SetClient(mockK8sObjTracker)
}
