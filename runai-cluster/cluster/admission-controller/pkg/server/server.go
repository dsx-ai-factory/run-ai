/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package server

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/handlers"

	"github.com/dyson/certman"
)

const (
	defaultReadTimeout  = 10
	defaultWriteTimeout = 10
	maxHeaderBytes      = 1 << 20 // 1048576
)

type Server struct {
	minAllocationEnforcementBytes  int
	sharedDirectoryHostPath        string
	mpsDirectoryHostPath           string
	coreDebugLogsDirectoryHostPath string
	coreDebugLevel                 string
	coreDebugFile                  string
	coreAnalytics                  bool
	isKubevirtEnabledCluster       bool
	scheduleWorkflowPods           bool
	specificClusterType            string
	overProvisioningRoute          bool
}

func NewServer() *Server {
	server := Server{}
	flag.IntVar(&server.minAllocationEnforcementBytes,
		"min-allocation-enforcement-bytes", 0, "Enforcement for fractional GPU jobs")
	flag.StringVar(&server.sharedDirectoryHostPath,
		"runai-shared-directory-host-path", "", "Shared directory host path for fractional GPU jobs")
	flag.StringVar(&server.mpsDirectoryHostPath,
		"runai-mps-directory-host-path", "", "Shared directory host path for mps GPU jobs")
	flag.StringVar(&server.coreDebugLogsDirectoryHostPath,
		"runai-core-debug-logs-directory-host-path", "", "Host path for core debug logs")
	flag.StringVar(&server.coreDebugLevel,
		"runai-core-debug-level", "", "Debug level of core components")
	flag.StringVar(&server.coreDebugFile,
		"runai-core-debug-file", "", "Log file path for core debugging")
	flag.BoolVar(&server.coreAnalytics,
		"runai-core-analytics", false, "Enable core analytics module")
	flag.BoolVar(&server.isKubevirtEnabledCluster,
		"runai-kubevirt-enabled-cluster", false, "Is Kubevirt enabled in the cluster")
	flag.BoolVar(&server.scheduleWorkflowPods,
		"runai-schedule-workflow-pods", false, "Should schedule workflow pods")
	flag.StringVar(&server.specificClusterType,
		"specific-cluster-type", "", "Specifies the cluster type where the operator is running")
	flag.BoolVar(&server.overProvisioningRoute,
		"overprovisioning-route", false,
		"Specifies if route capability of over provisioning is enabled")
	flag.Parse()

	return &server
}

func initCertWatcher() (*certman.CertMan, error) {
	cm, err := certman.New("/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key")
	if err != nil {
		return nil, err
	}
	cm.Logger(log.Default())
	if err := cm.Watch(); err != nil {
		return nil, err
	}
	return cm, nil
}

func (server *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	requestHandler := handlers.NewRequestHandler(
		ctx,
		server.minAllocationEnforcementBytes,
		server.sharedDirectoryHostPath,
		server.mpsDirectoryHostPath,
		server.coreDebugLogsDirectoryHostPath,
		server.coreDebugLevel,
		server.coreDebugFile,
		server.coreAnalytics,
		server.isKubevirtEnabledCluster,
		server.scheduleWorkflowPods,
		server.specificClusterType,
		server.overProvisioningRoute,
	)
	// Pod handlers
	mux.HandleFunc("/mutate-pod", requestHandler.HandleMutatePod)

	// Pod controller handlers
	mux.HandleFunc("/mutate-pod-controller", requestHandler.HandleMutatePodController)

	readTimeout, _ := strconv.Atoi(os.Getenv("WEBHOOK_SERVER_READ_TIMEOUT"))
	if readTimeout == 0 {
		readTimeout = defaultReadTimeout
	}

	writeTimeout, _ := strconv.Atoi(os.Getenv("WEBHOOK_SERVER_WRITE_TIMEOUT"))
	if writeTimeout == 0 {
		writeTimeout = defaultWriteTimeout
	}

	cm, err := initCertWatcher()
	if err != nil {
		return err
	}

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    time.Duration(readTimeout) * time.Second,
		WriteTimeout:   time.Duration(writeTimeout) * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
		TLSConfig: &tls.Config{
			GetCertificate: cm.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	return s.ListenAndServeTLS("", "")
}
