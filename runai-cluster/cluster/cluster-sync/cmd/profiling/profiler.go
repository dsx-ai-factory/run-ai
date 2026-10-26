/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package profiling

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"
)

// To use in cluster:
// brew install graphviz
// kubectl port-forward -n runai deploy/cluster-sync 8182
// go tool pprof -http localhost:8888 http://localhost:8182/debug/pprof/profile
// -> and then wait a little bit and it will open a window in the browser

func RegisterProfiler(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
