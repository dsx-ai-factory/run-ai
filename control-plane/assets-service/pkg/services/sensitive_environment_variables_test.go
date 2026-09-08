/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sensitive environment variable response sanitization", func() {
	It("omits HF and NGC literal values without mutating the source slice", func() {
		hfName, hfValue := "HF_TOKEN", "synthetic-hf-value"
		ngcName, ngcValue := "NGC_API_KEY", "synthetic-ngc-value"
		normalName, normalValue := "NORMAL_VAR", "visible"
		source := []api.EnvironmentVariableOfAsset{
			{Name: &hfName, Value: &hfValue},
			{Name: &ngcName, Value: &ngcValue},
			{Name: &normalName, Value: &normalValue},
		}

		redacted := redactSensitiveEnvironmentVariables(&source)

		Expect(*redacted).To(HaveLen(3))
		Expect((*redacted)[0].Name).To(Equal(&hfName))
		Expect((*redacted)[0].Value).To(BeNil())
		Expect((*redacted)[1].Value).To(BeNil())
		Expect((*redacted)[2].Value).To(Equal(&normalValue))
		Expect(source[0].Value).To(Equal(&hfValue))
		Expect(source[1].Value).To(Equal(&ngcValue))
	})

	It("preserves credential references", func() {
		name := "HF_TOKEN"
		credential := &api.EnvironmentVariableCredential{}
		source := []api.EnvironmentVariableOfAsset{{Name: &name, Credential: credential}}

		redacted := redactSensitiveEnvironmentVariables(&source)

		Expect((*redacted)[0].Credential).To(BeIdenticalTo(credential))
	})

	It("returns redacted copies for V2 response parameters", func() {
		name, value := "HF_TOKEN", "synthetic-hf-value"
		source := []api.EnvironmentVariableOfAsset{{Name: &name, Value: &value}}

		inference := redactInferenceSpecificRunParamsSensitiveValues(&api.InferenceSpecificRunParams{EnvironmentVariables: &source})
		workspace := redactWorkspaceSpecificRunParamsSensitiveValues(&api.WorkspaceSpecificRunParams{EnvironmentVariables: &source})
		training := redactTrainingSpecificRunParamsSensitiveValues(&api.TrainingSpecificRunParams{EnvironmentVariables: &source})

		Expect((*inference.EnvironmentVariables)[0].Value).To(BeNil())
		Expect((*workspace.EnvironmentVariables)[0].Value).To(BeNil())
		Expect((*training.EnvironmentVariables)[0].Value).To(BeNil())
		Expect(source[0].Value).To(Equal(&value))
	})

	It("redacts worker, master, and rendered workload representations", func() {
		hfName, hfValue := "HF_TOKEN", "synthetic-hf-value"
		ngcName, ngcValue := "NGC_API_KEY", "synthetic-ngc-value"
		workerVariables := []api.EnvironmentVariableOfAsset{{Name: &hfName, Value: &hfValue}}
		masterVariables := []api.EnvironmentVariableOfAsset{{Name: &ngcName, Value: &ngcValue}}
		rendered := map[string]interface{}{
			"spec": map[string]interface{}{
				"environment": map[string]interface{}{
					"items": map[string]interface{}{
						"NGC_API_KEY": map[string]interface{}{"value": "synthetic-dictionary-value"},
						"NORMAL_VAR":  map[string]interface{}{"value": "visible"},
					},
				},
				"containers": []interface{}{map[string]interface{}{
					"env": []interface{}{map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-rendered-value"}},
				}},
			},
		}
		workload := &dbmodel.GenericWorkload{
			Spec: api.SpecificRunInfoFields{
				SpecificEnv: &api.SpecificRunParams{EnvironmentVariables: &workerVariables},
				Distributed: &api.InfoDistributed{Master: &api.InfoDistMaster{
					SpecificEnv: &api.SpecificRunParams{EnvironmentVariables: &masterVariables},
				}},
			},
			Workload: rendered,
		}

		redactGenericWorkloadSensitiveValues(workload)

		Expect((*workload.Spec.SpecificEnv.EnvironmentVariables)[0].Value).To(BeNil())
		Expect((*workload.Spec.Distributed.Master.SpecificEnv.EnvironmentVariables)[0].Value).To(BeNil())
		responseEnv := workload.Workload["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})
		Expect(responseEnv[0].(map[string]interface{})).NotTo(HaveKey("value"))
		responseItems := workload.Workload["spec"].(map[string]interface{})["environment"].(map[string]interface{})["items"].(map[string]interface{})
		Expect(responseItems["NGC_API_KEY"].(map[string]interface{})).NotTo(HaveKey("value"))
		Expect(responseItems["NORMAL_VAR"].(map[string]interface{})["value"]).To(Equal("visible"))
		Expect(workerVariables[0].Value).To(Equal(&hfValue))
		Expect(masterVariables[0].Value).To(Equal(&ngcValue))
		storedEnv := rendered["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["env"].([]interface{})
		Expect(storedEnv[0].(map[string]interface{})["value"]).To(Equal("synthetic-rendered-value"))
		storedItems := rendered["spec"].(map[string]interface{})["environment"].(map[string]interface{})["items"].(map[string]interface{})
		Expect(storedItems["NGC_API_KEY"].(map[string]interface{})["value"]).To(Equal("synthetic-dictionary-value"))
	})
})
