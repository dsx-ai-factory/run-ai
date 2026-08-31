/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

const tsdbDir = "/prometheus/data"

func createBlocksFromOpenmetrics(w http.ResponseWriter, req *http.Request) {
	content, err := bodyToString(req.Body)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpFile, err := createTempFile(content)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())

	absTmpFileName, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := exec.Command("/bin/promtool", "tsdb", "create-blocks-from", "openmetrics", absTmpFileName, tsdbDir).Output() //nolint:gosec,noctx // G204: testing- no security hazard; integration test helper
	if err != nil {
		handleError(w, fmt.Sprintf("failed to run promtool: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

func handleError(w http.ResponseWriter, errMsg string, statusCode int) {
	logging.Errorf("%s", errMsg)

	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(errMsg))
}

func main() {
	port := os.Getenv("AGENT_PORT")
	if port == "" {
		port = "8090"
	}

	http.HandleFunc("/tsdb/create-blocks-from/openmetrics", createBlocksFromOpenmetrics)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		ReadHeaderTimeout: 2 * time.Minute,
	}
	logging.Infof("starts serving on port %s", port)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
