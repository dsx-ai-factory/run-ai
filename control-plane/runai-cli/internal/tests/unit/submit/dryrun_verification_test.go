/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	submitcmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/submit"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	assetsapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

const (
	IgnoreFolder = "ignore_folder"
)

var _ = Describe("Submit Test Dry Run", Ordered, func() {
	var (
		configPath           string
		err                  error
		envMap               map[string]string
		mockDatasourceLister *cpmocks.MockDatasourceDescribeLister
		projectList          = []string{"aidot-testing", "amplify", "internal-playground", "ltchapmi", "picasso-services", "rcp-ge-memuller", "test", "proj-defa", "yuan"}
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
		clients.MockProjectWithList(mockCtrl, projectList)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Training{}, trainingList)
		mockDatasourceLister = cpmocks.NewMockDatasourceDescribeLister(mockCtrl)
		datasource.SetClient(mockDatasourceLister)

		err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
		if err != nil {
			Fail("authentication failed, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("authentication failed, error: " + err.Error())
		}
	})

	AfterAll(func() {
		datasource.SetClient(nil)
	})

	BeforeEach(func() {
		submitcmd.SubmitSpec.Clear()
	})

	Context("Submit dry run validation", func() {
		It("attaches a datasource to the master pod", func() {
			mockDatasourceLister.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&assetsapi.DatasourceListResponse{
				Entries: []assetsapi.DatasourceListResponseEntry{{
					Meta: assetsapi.AssetMeta{Id: "master-datasource-id", Name: "master-data"},
				}},
			}, nil)

			output, err := cli.RunCommand("mpi submit -i ubuntu -p p301 --master-datasource type=pvc,name=master-data --dry-run", envMap)
			Expect(err).To(BeNil())

			distRequest := openapi.DistributedCreationRequest{}
			Expect(json.Unmarshal([]byte(output), &distRequest)).To(Succeed())
			masterAssets := distRequest.GetMasterAssets()
			Expect(masterAssets.GetDatasources()).To(Equal([]string{"master-datasource-id"}))
			Expect(distRequest.HasAssets()).To(BeFalse())
			Expect(distRequest.GetMasterSpecSameAsWorker()).To(BeFalse())
			Expect(distRequest.HasMasterSpec()).To(BeTrue())
		})

		It("exposes master-datasource for the other distributed training frameworks", func() {
			for _, framework := range []string{"pytorch", "tensorflow", "xgboost"} {
				output, err := cli.RunCommand(framework+" submit --help", envMap)
				Expect(err).To(BeNil(), framework)
				Expect(output).To(ContainSubstring("--master-datasource"), framework)
			}
		})

		It("rejects master-datasource when no master is requested", func() {
			for _, framework := range []string{"pytorch", "tensorflow"} {
				_, err := cli.RunCommand(framework+" submit -i ubuntu -p p301 --no-master --master-datasource type=pvc,name=master-data --dry-run", envMap)
				Expect(err).To(MatchError(ContainSubstring("if any flags in the group [no-master master-datasource] are set none of the others can be")), framework)
			}
		})

		It("verify mpi master master spec same as worker is true", func() {
			output, err := cli.RunCommand("mpi submit -i ubuntu -p p301 --dry-run", envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.Name).To(Equal("mpi"))
			Expect(distRequest.HasSpec()).To(Equal(true))
			Expect(distRequest.GetMasterSpecSameAsWorker()).To(Equal(true))
			Expect(distRequest.HasMasterSpec()).To(Equal(false))
		})

		It("verify xgboost master master spec same as worker is true", func() {
			output, err := cli.RunCommand("xgboost submit -i ubuntu -p p301 --dry-run", envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.Name).To(Equal("xgboost"))
			Expect(distRequest.HasSpec()).To(Equal(true))
			Expect(distRequest.GetMasterSpecSameAsWorker()).To(Equal(true))
			Expect(distRequest.HasMasterSpec()).To(Equal(false))
		})

		It("verify mpi worker and master has only one command annotation", func() {
			commandStr := "mpi submit -i ubuntu -p p301 --master-environment-variable a=b --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.Name).To(Equal("mpi"))
			Expect(distRequest.HasSpec()).To(Equal(true))
			Expect(distRequest.GetMasterSpecSameAsWorker()).To(Equal(false))
			Expect(distRequest.HasMasterSpec()).To(Equal(true))

			Expect(len(distRequest.GetSpec().Annotations)).To(Equal(1))
			Expect(distRequest.GetSpec().Annotations[0].GetName()).To(Equal(spec.CommandAnnotationKey))
			Expect(distRequest.GetSpec().Annotations[0].GetValue()).To(ContainSubstring(commandStr))
			Expect(distRequest.GetSpec().Annotations).To(Equal(distRequest.GetMasterSpec().Annotations))
		})

		Context("JAX verification", func() {
			It("help doesn't expose master container flags", func() {
				commandStr := "jax submit -h"
				output, err := cli.RunCommand(commandStr, envMap)
				Expect(err).To(BeNil())

				Expect(output).To(HavePrefix("submit a jax training workload"))
				Expect(output).ToNot(ContainSubstring("--master-command"))
				Expect(output).ToNot(ContainSubstring("--master-datasource"))
			})

			It("verify jax submit has only worker spec", func() {
				commandStr := "jax submit -i ubuntu -p p301 --dry-run"
				output, err := cli.RunCommand(commandStr, envMap)
				Expect(err).To(BeNil())
				distRequest := openapi.DistributedCreationRequest{}
				err = json.Unmarshal([]byte(output), &distRequest)
				Expect(err).To(BeNil())
				Expect(distRequest.Name).To(Equal("jax"))
				Expect(distRequest.HasSpec()).To(Equal(true))
				Expect(distRequest.GetMasterSpecSameAsWorker()).To(Equal(false))
				Expect(distRequest.HasMasterSpec()).To(Equal(false))
			})
		})

		It("verify submit with backoff limit 0 set spec with limit 0", func() {
			commandStr := "mpi submit -i ubuntu -p p301 --backoff-limit 0 --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.Name).To(Equal("mpi"))
			Expect(distRequest.HasSpec()).To(Equal(true))
			spec := distRequest.GetSpec()
			Expect(spec.GetBackoffLimit()).To(Equal(int64(0)))
		})

		It("verify submit with secret volume", func() {
			name, path, defaultMode := "Secret", "MountPath", "0666"
			commandStr := fmt.Sprintf("mpi submit -i ubuntu -p p301 --secret-volume=name=%s,path=%s,default-mode=%s --dry-run", name, path, defaultMode)
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.HasSpec()).To(Equal(true))
			spec := distRequest.GetSpec()
			secretVolume := spec.GetStorage().SecretVolume
			Expect(secretVolume[0].Secret.Get()).To(Equal(&name))
			Expect(secretVolume[0].MountPath.Get()).To(Equal(&path))
			Expect(secretVolume[0].DefaultMode.Get()).To(Equal(&defaultMode))
		})

		It("verify submit with master args and command", func() {
			commandStr := "mpi submit -i ubuntu -p p301 --master-command python --master-args bfoo --command --dry-run -- python -d baa"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.Name).To(Equal("mpi"))
			Expect(distRequest.HasSpec()).To(Equal(true))
			spec := distRequest.GetSpec()
			Expect(spec.GetCommand()).To(Equal("python -d baa"))
			master := distRequest.GetMasterSpec()
			Expect(master.GetCommand()).To(Equal("python"))
			Expect(master.GetArgs()).To(Equal("bfoo"))
		})

		It("verify submit with restart policy", func() {
			commandStr := "mpi submit test -i ubuntu --restart-policy Always -p test --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.HasSpec()).To(Equal(true))
			spec := distRequest.GetSpec()
			Expect(spec.GetRestartPolicy()).To(Equal(openapi.RESTARTPOLICY_ALWAYS))
		})

		It("verify submit with restart policy never with backoff limit will fail", func() {
			commandStr := "mpi submit test -i ubuntu --restart-policy Never --backoff-limit 1 -p test --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(Not(BeNil()))
			Expect(output).To(ContainSubstring("invalid input, cannot set backoff limit when restart policy is Never\n"))

		})

		It("verify submit with restart policy with master spec", func() {
			commandStr := "mpi submit test -i ubuntu --restart-policy Never --master-restart-policy Always -p test --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.HasSpec()).To(Equal(true))
			spec := distRequest.GetSpec()
			Expect(spec.GetRestartPolicy()).To(Equal(openapi.RESTARTPOLICY_NEVER))
			master := distRequest.GetMasterSpec()
			Expect(master.GetRestartPolicy()).To(Equal(openapi.RESTARTPOLICY_ALWAYS))
		})

		It("verify submit with field ref", func() {
			fieldRefName := "testname"
			fieldRefValue := "testvalue"
			commandStr := fmt.Sprintf("mpi submit test -i ubuntu --env-pod-field-ref %s=%s -p test --dry-run", fieldRefName, fieldRefValue)
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.HasSpec()).To(Equal(true))
			spec := distRequest.GetSpec()
			podFieldRef := spec.GetEnvironmentVariables()[0]
			Expect(*podFieldRef.Name.Get()).To(Equal(fieldRefName))
			Expect(podFieldRef.PodFieldRef.Get().GetPath()).To(Equal(fieldRefValue))
		})

		It("verify args value with different quotes set", func() {
			commandStr := fmt.Sprintf("mpi submit test1 -p test -i ubuntu --dry-run -- 'test --bla=\"test2 desc\"'")
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			workspace1 := openapi.Workspace{}
			err = json.Unmarshal([]byte(output), &workspace1)
			Expect(err).To(BeNil())
			Expect(workspace1.HasSpec()).To(Equal(true))

			commandStr = fmt.Sprintf("mpi submit test2 -p test -i ubuntu --dry-run -- \"test --bla='test2 desc'\"")
			output, err = cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			workspace2 := openapi.Workspace{}
			err = json.Unmarshal([]byte(output), &workspace2)
			Expect(err).To(BeNil())
			Expect(workspace2.HasSpec()).To(Equal(true))

			Expect(*workspace2.Spec.Args.Get()).To(Equal(*workspace1.Spec.Args.Get()))
			Expect(*workspace2.Spec.Annotations[0].Value.Get()).ToNot(Equal(*workspace1.Spec.Annotations[0].Value.Get()))
			Expect(*workspace2.Spec.Args.Get()).To(Equal("\"test --bla=\"test2 desc\"\""))
		})
	})

	Context("Submit dry run validation - Backward compatibility check", func() {
		It("verify submit worker gpu - master no gpu set", func() {
			commandStr := "submit-dist mpi mpi-interactive --master-gpu=0 --workers=2 -g 8 -i nvcr.io/nvidia/nemo:24.07 -p test --master-args /bin/bash --dry-run --command -- /usr/sbin/sshd"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(consts.WarningSubmitMpuWithGpu))
			output = removeJsonStringPrefix(output)
			distRequest := openapi.DistributedCreationRequest{}
			err = json.Unmarshal([]byte(output), &distRequest)
			Expect(err).To(BeNil())
			Expect(distRequest.Name).To(Equal("mpi-interactive"))
			Expect(distRequest.HasSpec()).To(Equal(true))
			spec := distRequest.GetSpec()
			Expect(spec.GetCompute().GpuPortionRequest.IsSet()).To(BeFalse())
			Expect(spec.GetCompute().GpuDevicesRequest.IsSet()).To(BeTrue())
			gpuDevices := spec.GetCompute().GpuDevicesRequest.Get()
			Expect(gpuDevices).ToNot(BeNil())
			Expect(*gpuDevices).To(Equal(int32(8)))
			master := distRequest.GetMasterSpec()
			Expect(master.GetCompute().GpuPortionRequest.IsSet()).To(BeFalse())
			Expect(master.GetCompute().GpuDevicesRequest.IsSet()).To(BeFalse())
		})

		It("verify submit gpu and master gpu - request", func() {
			commandStr := "submit-dist mpi job-mpi-name -i ubuntu -p p301 --gpu 789 --master-gpu 123 --master-command python --master-args bfoo --command --dry-run -- python -d baa"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(consts.ErrorSubmitMpiWithGpu))
		})

		It("verify submit with gpu and master gpu - master portion, worker request", func() {
			commandStr := "submit-dist mpi job-mpi-name -i ubuntu -p p301 --gpu 2 --master-gpu 0.321 --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(consts.ErrorSubmitMpiWithGpu))
		})

		It("verify submit with gpu and master gpu - master request, worker portion", func() {
			commandStr := "submit-dist mpi job-mpi-name -i ubuntu -p p301 --gpu 0.678 --master-gpu 99 --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(consts.ErrorSubmitMpiWithGpu))
		})

		It("verify submit with name, gpu and master gpu - portion", func() {
			commandStr := "submit-dist mpi --name job-mpi-with-name -i ubuntu -p p301 --gpu 0.123 --master-gpu 0.321 --dry-run"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(consts.ErrorSubmitMpiWithGpu))
		})

		It("load output.txt file from ignore_folder - validation only, will not fail tests", func() {
			Skip("not needed in day to day")
			homePath, err := cli.GetHomePath()
			if err != nil {
				Fail("failed to get pwd")
			}
			outputFile := GetTestFile(homePath, "output.txt")
			file, err := os.Open(outputFile)
			if err != nil {
				return
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				submitcmd.SubmitSpec.Clear()
				output, err := cli.RunCommand(line, envMap)
				output = removeJsonStringPrefix(output)
				if err != nil {
					log.Printf("error: %s, failed to run line: %s", err.Error(), line)
					continue
				}

				if !strings.HasPrefix(output, "{") {
					_, post, found := strings.Cut(output, "\n")
					if found {
						output = post
					}
				}
				distRequest := openapi.DistributedCreationRequest{}
				err = json.Unmarshal([]byte(output), &distRequest)
				Expect(err).To(BeNil())
				Expect(distRequest.HasSpec()).To(Equal(true))
			}
		})
	})
})

func removeJsonStringPrefix(output string) string {
	return output[strings.Index(output, "{"):]
}

func GetTestFile(homePath, testFileName string) string {
	return filepath.Join(homePath, IgnoreFolder, testFileName)
}
