/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/pvc"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/golang/mock/gomock"
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
