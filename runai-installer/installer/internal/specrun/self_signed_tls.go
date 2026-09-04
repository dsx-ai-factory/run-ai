/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package specrun

import (
	"os"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/cert"
)

func prepareSelfSignedTLS(s *spec.Spec, sink orchestrator2.EventSink) error {
	if !s.SelfSignedTLSEnabled() || s.DryRun {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	outputDir := filepath.Join(cwd, cert.SelfSignedDirectoryName)
	if err := sink.HandleEvent(orchestrator2.Event{
		Kind:          orchestrator2.EventKindSelfSignedTLS,
		SelfSignedTLS: orchestrator2.NewSelfSignedTLSWarning(outputDir),
	}); err != nil {
		return err
	}
	generated, err := cert.EnsureSelfSignedCertificateSet(cwd, s.SelfSignedTLSDNSNames())
	if err != nil {
		return err
	}
	s.TLS = &spec.TLS{
		CertPath:    generated.CertPath,
		KeyPath:     generated.KeyPath,
		CAChainPath: generated.CACertPath,
	}
	s.SelfSignedTLS.OutputDir = generated.OutputDir
	return nil
}
