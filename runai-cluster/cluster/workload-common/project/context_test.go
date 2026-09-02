/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type TestCase struct {
	ns            string
	usage         string
	expectedError string
}

func TestSetProjectInContextFromNamespace(t *testing.T) {

	k8sClient := fake.NewClientBuilder().WithRuntimeObjects(EmulateProjNs("runai-team-a", "team-a"), EmulateRunaiNs()).Build()

	testCases := []TestCase{
		{
			ns:            "runai-team-a",
			usage:         "Template",
			expectedError: "",
		},
		{
			ns:            "",
			usage:         "Template",
			expectedError: "",
		},
		{
			ns:            "runai-team-a",
			usage:         "Submit",
			expectedError: "",
		},
		{
			ns:            "",
			usage:         "Submit",
			expectedError: "xxx",
		},
	}

	for _, testCase := range testCases {
		om := metav1.ObjectMeta{
			Name:      "test",
			Namespace: testCase.ns,
		}

		_, err := SetProjectInContextFromNamespace(context.Background(), k8sClient, om.Namespace, om.Name, testCase.usage)

		if testCase.expectedError != "" {
			assert.Error(t, err, "ns '%s'/usage %s", testCase.ns, testCase.usage)
		} else {
			assert.NoError(t, err, "ns '%s'/usage %s", testCase.ns, testCase.usage)
		}

	}
}
