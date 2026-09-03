/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/cluster"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/golang/mock/gomock"
)

func MockClusterStatus(mockCtrl *gomock.Controller) {
	status := mocks.NewMockClusterStatusClient(mockCtrl)
	status.EXPECT().GetClusterStatus(gomock.Any()).Return(nil, fmt.Errorf("bla"))
	status.EXPECT().GetClusterStatus(gomock.Any()).Return(nil, nil).AnyTimes()
	status.EXPECT().GetWorkloadStatus(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("bla")).AnyTimes()
	cluster.SetClient(status)
}

func MockClusterStatusTimedOut(mockCtrl *gomock.Controller) {
	status := mocks.NewMockClusterStatusClient(mockCtrl)
	status.EXPECT().GetClusterStatus(gomock.Any()).Return(nil, &url.Error{
		Op:  "Get",
		URL: "https://timed.out.url.com",
		Err: context.DeadlineExceeded,
	})
	cluster.SetClient(status)
}
