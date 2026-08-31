/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package profiling

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func RegisterProfiler(port string) {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	pprof.Register(router)
	err = router.Run(fmt.Sprintf(":%v", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
