/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/inference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
)

const (
	TestFilePath = "inference"
)

var workloadList = []string{"infw-minimal", "infw-compute-full", "inference-prefix"}

func ResetRunaiCommand() {
	cmd := inference.Cmd
	for _, c := range cmd.Commands() {
		cmd.RemoveCommand(c)
	}
	cmdfactory.WorkloadsCmdInit(cmd, api.Knative, api.Inference)
}

var _ = Describe("Inference Update Test Dry Run", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())

		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Inference{}, workloadList)

		err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
		if err != nil {
			Fail("authentication failed, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("authentication failed, error: " + err.Error())
		}
	})

	BeforeEach(func() {
		ResetRunaiCommand()
	})

	Context("Update dry run validation", func() {
		It("verify update with basic flags", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --image=new-img --image-pull-policy=Always --working-dir=/a/b/c --activation-replicas=1 --concurrency-hard-limit=0 --initial-replicas=2 --initialization-timeout-seconds=4 --request-timeout-seconds=11 --max-replicas=5 --metric=metric --metric-threshold=6 --metric-threshold-percentage=70 --min-replicas=1 --scale-down-delay-seconds=8 --scale-to-zero-retention-seconds=9 --dry-run", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			sanitizedOutput := sanitizeJsonOutput(output)
			var jsonOutput map[string]interface{}
			err = json.Unmarshal([]byte(sanitizedOutput), &jsonOutput)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to unmarshal JSON: %s", sanitizedOutput))

			Expect(jsonOutput).To(HaveKeyWithValue("image", "new-img"))
			Expect(jsonOutput).To(HaveKeyWithValue("imagePullPolicy", "Always"))
			Expect(jsonOutput).To(HaveKeyWithValue("workingDir", "/a/b/c"))

			Expect(jsonOutput).To(HaveKey("autoscaling"))
			autoscalingSpec, ok := jsonOutput["autoscaling"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "autoscaling field is not a map")
			Expect(autoscalingSpec).To(HaveKeyWithValue("maxReplicas", float64(5)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("minReplicas", float64(1)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("initialReplicas", float64(2)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("activationReplicas", float64(1)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("metric", "metric"))
			Expect(autoscalingSpec).To(HaveKeyWithValue("metricThreshold", float64(6)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("metricThresholdPercentage", float64(70)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("concurrencyHardLimit", float64(0)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("scaleDownDelaySeconds", float64(8)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("scaleToZeroRetentionSeconds", float64(9)))
			Expect(autoscalingSpec).ToNot(HaveKey("initializationTimeoutSeconds"))

			servingConfiguration, ok := jsonOutput["servingConfiguration"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "servingConfiguration field is not a map")
			Expect(servingConfiguration).To(HaveKeyWithValue("initializationTimeoutSeconds", float64(4)))
			Expect(servingConfiguration).To(HaveKeyWithValue("requestTimeoutSeconds", float64(11)))
		})

		It("verify update with compute fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --cpu-core-request=0.5 --cpu-core-limit=1.0 --cpu-memory-request=1G --cpu-memory-limit=2G --gpu-devices-request=1 --gpu-memory-request=8G --gpu-memory-limit=16G --large-shm --dry-run", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			sanitizedOutput := sanitizeJsonOutput(output)
			var jsonOutput map[string]interface{}
			err = json.Unmarshal([]byte(sanitizedOutput), &jsonOutput)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to unmarshal JSON: %s", sanitizedOutput))

			Expect(jsonOutput).To(HaveKey("compute"))
			computeSpec, ok := jsonOutput["compute"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "compute field is not a map")
			Expect(computeSpec).To(HaveKeyWithValue("cpuCoreRequest", 0.5))
			Expect(computeSpec).To(HaveKeyWithValue("cpuCoreLimit", 1.0))
			Expect(computeSpec).To(HaveKeyWithValue("cpuMemoryRequest", "1G"))
			Expect(computeSpec).To(HaveKeyWithValue("cpuMemoryLimit", "2G"))
			Expect(computeSpec).To(HaveKeyWithValue("gpuDevicesRequest", float64(1)))
			Expect(computeSpec).To(HaveKeyWithValue("gpuMemoryRequest", "8G"))
			Expect(computeSpec).To(HaveKeyWithValue("gpuMemoryLimit", "16G"))
			Expect(computeSpec).To(HaveKeyWithValue("largeShmRequest", true))
		})

		It("verify update with node selection fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --node-pools=pool1 --node-pools=pool2 --dry-run", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			sanitizedOutput := sanitizeJsonOutput(output)
			var jsonOutput map[string]interface{}
			err = json.Unmarshal([]byte(sanitizedOutput), &jsonOutput)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to unmarshal JSON: %s", sanitizedOutput))

			Expect(jsonOutput).To(HaveKey("nodePools"))
			nodePools, ok := jsonOutput["nodePools"].([]interface{})
			Expect(ok).To(BeTrue(), "nodePools field is not an array")
			Expect(nodePools).To(ConsistOf("pool1", "pool2"))
		})

		It("verify update with combined fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --image=new-img --cpu-core-request=0.5 --gpu-devices-request=1 --node-pools=pool1 --min-replicas=1 --max-replicas=5 --dry-run", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			sanitizedOutput := sanitizeJsonOutput(output)
			var jsonOutput map[string]interface{}
			err = json.Unmarshal([]byte(sanitizedOutput), &jsonOutput)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to unmarshal JSON: %s", sanitizedOutput))

			Expect(jsonOutput).To(HaveKeyWithValue("image", "new-img"))

			Expect(jsonOutput).To(HaveKey("compute"))
			computeSpec, ok := jsonOutput["compute"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "compute field is not a map")
			Expect(computeSpec).To(HaveKeyWithValue("cpuCoreRequest", 0.5))
			Expect(computeSpec).To(HaveKeyWithValue("gpuDevicesRequest", float64(1)))

			Expect(jsonOutput).To(HaveKey("nodePools"))
			nodePools, ok := jsonOutput["nodePools"].([]interface{})
			Expect(ok).To(BeTrue(), "nodePools field is not an array")
			Expect(nodePools).To(ConsistOf("pool1"))

			Expect(jsonOutput).To(HaveKey("autoscaling"))
			autoscalingSpec, ok := jsonOutput["autoscaling"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "autoscaling field is not a map")
			Expect(autoscalingSpec).To(HaveKeyWithValue("minReplicas", float64(1)))
			Expect(autoscalingSpec).To(HaveKeyWithValue("maxReplicas", float64(5)))
		})

		It("verify update with security fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --dry-run", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			sanitizedOutput := sanitizeJsonOutput(output)
			var jsonOutput map[string]interface{}
			err = json.Unmarshal([]byte(sanitizedOutput), &jsonOutput)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to unmarshal JSON: %s", sanitizedOutput))
			// Since no specific security flags are supported for update,
			// we might just check that the command runs and produces some valid JSON (which is already done).
			// Or, we could check for default security values if applicable.
			// For now, just ensuring it doesn't crash and produces JSON is the main goal after flag removal.
		})

		It("verify update with storage fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --dry-run", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			sanitizedOutput := sanitizeJsonOutput(output)
			var jsonOutput map[string]interface{}
			err = json.Unmarshal([]byte(sanitizedOutput), &jsonOutput)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to unmarshal JSON: %s", sanitizedOutput))
			// Similar to security, no specific storage flags are supported for update.
			// We mainly ensure the command runs and produces valid JSON.
		})

		It("verify update with network fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --dry-run", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			sanitizedOutput := sanitizeJsonOutput(output)
			var jsonOutput map[string]interface{}
			err = json.Unmarshal([]byte(sanitizedOutput), &jsonOutput)
			Expect(err).To(BeNil(), fmt.Sprintf("Failed to unmarshal JSON: %s", sanitizedOutput))
			// Similar to security & storage, no specific network flags for port/external-url are supported for update.
			// We mainly ensure the command runs and produces valid JSON.
		})
	})
})

func sanitizeJsonOutput(output string) string {
	firstBrace := strings.Index(output, "{")
	if firstBrace == -1 {
		return output
	}

	lastBrace := strings.LastIndex(output, "}")
	if lastBrace == -1 || lastBrace < firstBrace {
		return output
	}

	return output[firstBrace : lastBrace+1]
}
