/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"log"
	"os"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
)

const (
	ccInitdataAnnotation = "io.katacontainers.config.hypervisor.cc_init_data"
	ccInitdataPath       = "/etc/runai/cc-initdata/initdata.toml"
)

func (pm *PodMutator) mutateConfidentialInitdata(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	if pod.Labels[constants.ConfidentialComputingLabel] != "true" {
		return nil
	}

	if _, found := pod.Annotations[ccInitdataAnnotation]; found {
		return nil
	}

	document, err := os.ReadFile(ccInitdataPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: could not read initdata document: %v", err)
		}
		return nil
	}

	encoded, err := encodeInitdata(document)
	if err != nil {
		log.Printf("Warning: could not encode initdata document: %v", err)
		return nil
	}

	patch.AddAnnotationToObjectMeta(&pod.ObjectMeta, ccInitdataAnnotation, encoded)
	return nil
}

func encodeInitdata(document []byte) (string, error) {
	var compressed bytes.Buffer

	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(document); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(compressed.Bytes()), nil
}
