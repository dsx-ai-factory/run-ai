/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package backendclient

import (
	"sync"
)

type SyncHandler interface {
	HandleSyncEvent(waitGroup *sync.WaitGroup)
}
