/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_test

import (
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/utils"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/inference"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	TestFilePath                                     = "inference"
	ExpectedInferenceListWithProjectThatDoesNotExist = "project name: p300, not found\n"
	deletingMsg                                      = "Deleting workload '%s'...\n"
	updatingMsg                                      = "inference %s updated successfully\n"
	MissingProject                                   = "Error: project is missing. value is empty.\n"
	InferenceMinimalName                             = "infw-minimal"
	InferenceMinimalImageName                        = "minimal-image"
	ExpectSuccessMsg                                 = "Creating inference %s..."
	InferenceComputeFullName                         = "infw-compute-full"
	InferenceComputeFullNameImageName                = "compute-full-image"
	InferenceServingPort                             = "1234"
	InferencePrefixName                              = "inference-prefix"
	runAsUserError                                   = "Error: Invalid flags. if [run-as-user] flag is set non of the others [run-as-uid, run-as-gid, user-group-source, supplemental-groups] should be set\n"
)

var workloadList = []string{InferenceMinimalName, InferenceComputeFullName, InferencePrefixName}

var _ = Describe("Inference Help Output Test", Ordered, func() {
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
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
		ResetRunaiCommand()
	})

	Context("Inference Help Validation", func() {
		It("should show inference architectures including distributed in help output", func() {
			output, err := cli.RunCommand("inference --help", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Inference Architecture"))
			Expect(output).To(ContainSubstring("standard"))
			Expect(output).To(ContainSubstring("distributed"))
			Expect(output).To(ContainSubstring("Runs multiple coordinated inference processes across multiple nodes. Required for models too large to run on a single node."))
			Expect(output).To(ContainSubstring("Runs a single inference process on one node. Suitable for smaller models or simpler inference tasks."))
		})

		It("should show list command in inference general commands", func() {
			output, err := cli.RunCommand("inference --help", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Inference General Commands"))
			Expect(output).To(ContainSubstring("list"))
		})

		It("should show submit command as both shortcut and under standard", func() {
			output, err := cli.RunCommand("inference --help", envMap)
			Expect(err).To(BeNil())
			// Submit should appear as a standard inference shortcut
			Expect(output).To(ContainSubstring("submit"))
		})
	})
})

var _ = Describe("Inference Standard Submit Equivalence Test", Ordered, func() {
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
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Inference{}, workloadList)
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
		ResetRunaiCommand()
	})

	Context("Submit Command Equivalence", func() {
		It("should produce same spec with 'inference submit' and 'inference standard submit'", func() {
			// Submit using 'inference submit'
			_, err1 := cli.RunCommand(fmt.Sprintf(
				"inference submit %s -p p301 -i %s --serving-port 8080 --dry-run",
				InferenceMinimalName, InferenceMinimalImageName), envMap)
			Expect(err1).To(BeNil())

			spec1, err := clients.GetSpecByInferenceName(InferenceMinimalName)
			Expect(err).To(BeNil())

			// Reset for second submission
			ResetRunaiCommand()

			// Submit using 'inference standard submit'
			_, err2 := cli.RunCommand(fmt.Sprintf(
				"inference standard submit %s -p p301 -i %s --serving-port 8080 --dry-run",
				InferenceMinimalName, InferenceMinimalImageName), envMap)
			Expect(err2).To(BeNil())

			spec2, err := clients.GetSpecByInferenceName(InferenceMinimalName)
			Expect(err).To(BeNil())

			// Compare specs - they should be identical
			Expect(spec1.GetImage()).To(Equal(spec2.GetImage()))
			Expect(spec1.GetServingPort()).To(Equal(spec2.GetServingPort()))
		})

		It("should produce same spec with complex flags using both commands", func() {
			// Submit using 'inference submit' with complex flags
			_, err1 := cli.RunCommand(fmt.Sprintf(
				"inference submit %s -p p301 -i %s --serving-port 8080 -g 2 --cpu-core-request 4 --cpu-memory-request 8G --dry-run",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err1).To(BeNil())

			spec1, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())

			// Reset for second submission
			ResetRunaiCommand()

			// Submit using 'inference standard submit' with same flags
			_, err2 := cli.RunCommand(fmt.Sprintf(
				"inference standard submit %s -p p301 -i %s --serving-port 8080 -g 2 --cpu-core-request 4 --cpu-memory-request 8G --dry-run",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err2).To(BeNil())

			spec2, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())

			// Compare specs - they should be identical
			Expect(spec1.GetImage()).To(Equal(spec2.GetImage()))
			Expect(spec1.GetServingPort()).To(Equal(spec2.GetServingPort()))
			compute1 := spec1.GetCompute()
			compute2 := spec2.GetCompute()
			Expect(compute1.GetGpuDevicesRequest()).To(Equal(compute2.GetGpuDevicesRequest()))
			Expect(compute1.GetCpuCoreRequest()).To(Equal(compute2.GetCpuCoreRequest()))
			Expect(compute1.GetCpuMemoryRequest()).To(Equal(compute2.GetCpuMemoryRequest()))
		})
	})
})

var _ = Describe("Inference Test", Ordered, func() {
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
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Inference{}, workloadList)
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
		ResetRunaiCommand()
	})

	Context("Inference delete validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})
		It("inference delete no uuid or name", func() {
			output, err := cli.RunCommand("inference delete --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to delete workload. one of the attributes must be set name|uuid\n"))
		})

		It("inference delete both uuid and name", func() {
			output, err := cli.RunCommand("inference delete anyname --project=p301 --uuid=anothername", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to delete workload. one of the attributes must be set name|uuid\n"))
		})

		It("inference delete, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference delete --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(deletingMsg, clients.UuidExist)))
		})

		It("inference delete, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference delete --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to delete workload. generic error string\n"))
		})

		It("inference delete success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference delete %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(deletingMsg, clients.WorkloadName01)))
		})

		It("inference delete alias, failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference del %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(deletingMsg, clients.WorkloadName01)))
		})
	})

	Context("Inference submit prerequisite validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("inference submit - no project", func() {
			output, err := cli.RunCommand("inference submit", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix(MissingProject))
		})

		It("inference submit - no project", func() {
			output, err := cli.RunCommand("inference submit anyname", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix(MissingProject))
		})

		It("inference submit - project does not exist", func() {
			output, err := cli.RunCommand("inference submit infw01 --project=p300 --image=lies", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedInferenceListWithProjectThatDoesNotExist))
		})
	})

	Context("Inference submit validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("success naive", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s --serving-port=%s",
				InferenceMinimalName, InferenceMinimalImageName, InferenceServingPort), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceMinimalName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceMinimalName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceMinimalImageName))

			Expect(submitSpec.GetCreateHomeDir()).To(Equal(false))
			Expect(*submitSpec.ServingPort.Get().Container.Get()).To(Equal(int32(1234)))
		})

		It("success container", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
				"--command "+
				"--image-pull-policy=IfNotPresent "+
				"--working-dir=/test/path "+
				"--create-home-dir "+
				"--environment-variable=env1=value1 "+
				"--required-pod-topology-key=IS_REQ "+
				"-- this is what I need",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))

			Expect(submitSpec.GetCommand()).To(Equal("this is what I need"))
			Expect(submitSpec.GetImagePullPolicy()).To(Equal(openapi.IMAGEPULLPOLICY_IF_NOT_PRESENT))
			Expect(submitSpec.GetWorkingDir()).To(Equal("/test/path"))

			Expect(submitSpec.GetCreateHomeDir()).To(Equal(true))
			Expect(submitSpec.GetArgs()).To(Equal(""))

			pa := submitSpec.GetPodAffinity()
			Expect(pa.GetType()).To(Equal(openapi.PODAFFINITYTYPE_REQUIRED))
			Expect(pa.GetKey()).To(Equal("IS_REQ"))

			env := submitSpec.GetEnvironmentVariables()
			Expect(len(env)).To(Equal(1))
			Expect(env[0].GetName()).To(Equal("env1"))
			Expect(env[0].GetValue()).To(Equal("value1"))
		})

		It("success compute", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s --serving-port=%s "+
				"--cpu-memory-request=2G --cpu-memory-limit=1G "+
				"--cpu-core-limit=0.2 --cpu-core-request=0.3 "+
				"--gpu-devices-request=2 "+
				"--gpu-memory-limit=20M --gpu-memory-request=30M "+
				"--gpu-portion-limit=0.23 --gpu-portion-request=0.43 "+
				"--gpu-request-type=memory "+
				"--large-shm=true",
				InferenceComputeFullName, InferenceComputeFullNameImageName, InferenceServingPort), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))
			Expect(*submitSpec.Compute.Get()).ToNot(BeNil())

			compute := submitSpec.Compute.Get()
			Expect(*compute.CpuMemoryRequest.Get()).To(Equal("2G"))
			Expect(*compute.CpuMemoryLimit.Get()).To(Equal("1G"))
			Expect(*compute.GpuDevicesRequest.Get()).To(Equal(int32(2)))
			Expect(*compute.GpuMemoryRequest.Get()).To(Equal("30M"))
			Expect(*compute.GpuMemoryLimit.Get()).To(Equal("20M"))
			Expect(*compute.GpuPortionRequest.Get()).To(Equal(0.43))
			Expect(*compute.GpuPortionLimit.Get()).To(Equal(0.23))
			Expect(*compute.GpuRequestType.Get()).To(Equal(openapi.GpuRequestType("memory")))
			Expect(*compute.LargeShmRequest.Get()).To(Equal(true))
		})

		It("success general", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s --serving-port=%s "+
				"--annotation=ankey1=anvalue1 --annotation=ankey2=anvalue2 "+
				"--label=lkey1=lvalue1 --label=lkey2=lvalue2",
				InferenceComputeFullName, InferenceComputeFullNameImageName, InferenceServingPort), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(BeTrue())
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))
			Expect(submitSpec.Compute.IsSet()).To(BeFalse())

			Expect(submitSpec.Labels).ToNot(BeNil())
			Expect(len(submitSpec.Labels)).To(Equal(2))
			Expect(*submitSpec.Labels[0].Name.Get()).To(Equal("lkey1"))
			Expect(*submitSpec.Labels[0].Value.Get()).To(Equal("lvalue1"))
			Expect(*submitSpec.Labels[1].Name.Get()).To(Equal("lkey2"))
			Expect(*submitSpec.Labels[1].Value.Get()).To(Equal("lvalue2"))

			Expect(submitSpec.Annotations).ToNot(BeNil())
			Expect(len(submitSpec.Annotations)).To(Equal(3))
			Expect(*submitSpec.Annotations[0].Name.Get()).To(Equal("ankey1"))
			Expect(*submitSpec.Annotations[0].Value.Get()).To(Equal("anvalue1"))
			Expect(*submitSpec.Annotations[1].Name.Get()).To(Equal("ankey2"))
			Expect(*submitSpec.Annotations[1].Value.Get()).To(Equal("anvalue2"))
			Expect(*submitSpec.Annotations[2].Name.Get()).To(Equal("runai-submit-command"))
		})

		It("success storage", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
				"--configmap-map-volume=name=CONFIGMAP_NAME,path=PATH "+
				"--existing-pvc=claimname=CLAIM_NAME_EXIST,path=PATH_EXIST_PVC "+
				"--new-pvc=claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral "+
				"--git-sync=name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION "+
				"--host-path=path=PATH,mount=MOUNT,mount-propagation=HostToContainer,readwrite "+
				"--nfs=path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))
			Expect(submitSpec.Compute.IsSet()).To(BeFalse())

			Expect(submitSpec.Storage.IsSet()).To(Equal(true))
			storage := submitSpec.Storage.Get()

			Expect(len(storage.HostPath)).To(Equal(1))
			// path=PATH,mount=MOUNT,mount-propagation=HostToContainer,readwrite
			Expect(storage.HostPath[0].GetPath()).To(Equal("PATH"))
			Expect(storage.HostPath[0].GetMountPath()).To(Equal("MOUNT"))
			Expect(storage.HostPath[0].GetMountPropagation()).To(Equal(openapi.HOSTPATHMOUNTPROPAGATION_HOST_TO_CONTAINER))
			Expect(storage.HostPath[0].GetReadOnly()).To(Equal(false))

			Expect(len(storage.Nfs)).To(Equal(1))
			// path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite
			Expect(storage.Nfs[0].GetPath()).To(Equal("PATH"))
			Expect(storage.Nfs[0].GetServer()).To(Equal("SERVER"))
			Expect(storage.Nfs[0].GetMountPath()).To(Equal("MOUNT_PATH"))
			Expect(storage.Nfs[0].GetReadOnly()).To(Equal(false))

			Expect(len(storage.Pvc)).To(Equal(2))
			// claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral
			Expect(storage.Pvc[0].GetExistingPvc()).To(Equal(false))
			Expect(storage.Pvc[0].GetClaimName()).To(Equal("CLAIM_NAME"))
			Expect(storage.Pvc[0].GetPath()).To(Equal("PATH"))
			Expect(storage.Pvc[0].GetReadOnly()).To(Equal(true))
			Expect(storage.Pvc[0].GetEphemeral()).To(Equal(true))
			ci := storage.Pvc[0].GetClaimInfo()
			Expect(ci.GetSize()).To(Equal("SIZE"))
			Expect(ci.GetStorageClass()).To(Equal("STORAGE_CLASS"))
			am := ci.GetAccessModes()
			Expect(am.GetReadWriteOnce()).To(Equal(true))
			Expect(am.GetReadWriteMany()).To(Equal(true))
			Expect(am.GetReadOnlyMany()).To(Equal(true))

			// claimname=CLAIM_NAME_EXIST,path=PATH_EXIST_PVC
			Expect(storage.Pvc[1].GetExistingPvc()).To(Equal(true))
			Expect(storage.Pvc[1].GetClaimName()).To(Equal("CLAIM_NAME_EXIST"))
			Expect(storage.Pvc[1].GetPath()).To(Equal("PATH_EXIST_PVC"))
			Expect(storage.Pvc[1].GetReadOnly()).To(Equal(false))
			Expect(storage.Pvc[1].HasEphemeral()).To(Equal(false))
			Expect(storage.Pvc[1].HasClaimInfo()).To(Equal(false))

			Expect(len(storage.Git)).To(Equal(1))
			// name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION
			Expect(storage.Git[0].GetName()).To(Equal("NAME"))
			Expect(storage.Git[0].GetRepository()).To(Equal("REPO"))
			Expect(storage.Git[0].GetPath()).To(Equal("PATH"))
			Expect(storage.Git[0].GetPasswordSecret()).To(Equal("SECRET"))
			Expect(storage.Git[0].GetRevision()).To(Equal("REVISION"))

			Expect(len(storage.ConfigMapVolume)).To(Equal(1))
			// name=CONFIGMAP_NAME,path=PATH
			Expect(storage.ConfigMapVolume[0].GetConfigMap()).To(Equal("CONFIGMAP_NAME"))
			Expect(storage.ConfigMapVolume[0].GetMountPath()).To(Equal("PATH"))
		})

		It("success network", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
				"--port=service-type=NodePort,container=80,external=8080 "+
				"--external-url=container=9443,url=https://external.runai.com,authusers=user1:user2,authgroups=group1:group3",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))
			Expect(submitSpec.Compute.IsSet()).To(BeFalse())
			Expect(submitSpec.Storage.IsSet()).To(BeFalse())

			ports := submitSpec.GetPorts()
			Expect(len(ports)).To(Equal(1))
			Expect(ports[0].GetContainer()).To(Equal(int32(80)))
			Expect(ports[0].GetExternal()).To(Equal(int32(8080)))
			Expect(ports[0].GetServiceType()).To(Equal(openapi.PORTSERVICETYPE_NODE_PORT))
			exposedUrls := submitSpec.GetExposedUrls()
			Expect(len(exposedUrls)).To(Equal(1))
			Expect(exposedUrls[0].GetContainer()).To(Equal(int32(9443)))
			Expect(exposedUrls[0].GetUrl()).To(Equal("https://external.runai.com"))
			Expect(exposedUrls[0].GetAuthorizedUsers()).To(Equal([]string{"user1", "user2"}))
			Expect(exposedUrls[0].GetAuthorizedGroups()).To(Equal([]string{"group1", "group3"}))
		})

		It("success Inference", func() {

			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
				"--serving-port=container=1234,authorization-type=authorizedUsersOrGroups,authorized-users=user1:user2,authorized-groups=group1:group2,cluster-local-access-only,protocol=http "+
				"--initialization-timeout-seconds=10 "+ // stand alone flag
				"--request-timeout-seconds=11 "+ // stand alone flag
				"--activation-replicas=1 "+
				"--concurrency-hard-limit=2 "+
				"--initial-replicas=3 "+
				"--max-replicas=4 "+
				"--metric=myMetric "+
				"--metric-threshold=4 "+
				"--metric-threshold-percentage=4 "+
				"--min-replicas=4 "+
				"--scale-down-delay-seconds=4 "+
				"--scale-to-zero-retention-seconds=4",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))

			// autoscaling
			autoscaling := submitSpec.GetAutoscaling()
			Expect(autoscaling.GetActivationReplicas()).To(Equal(int32(1)))
			Expect(autoscaling.GetConcurrencyHardLimit()).To(Equal(int32(2)))
			Expect(autoscaling.GetInitialReplicas()).To(Equal(int32(3)))
			Expect(autoscaling.GetMaxReplicas()).To(Equal(int32(4)))
			Expect(autoscaling.GetMetric()).To(Equal("myMetric"))
			Expect(autoscaling.GetMetricThreshold()).To(Equal(int32(4)))
			Expect(autoscaling.GetMetricThresholdPercentage()).To(Equal(float32(4)))
			Expect(autoscaling.GetMinReplicas()).To(Equal(int32(4)))
			Expect(autoscaling.GetScaleDownDelaySeconds()).To(Equal(int32(4)))
			Expect(autoscaling.GetScaleToZeroRetentionSeconds()).To(Equal(int32(4)))

			// servingConfiguration
			servingConfiguration := submitSpec.GetServingConfiguration()
			Expect(servingConfiguration.GetInitializationTimeoutSeconds()).To(Equal(int32(10)))
			Expect(servingConfiguration.GetRequestTimeoutSeconds()).To(Equal(int32(11)))

			// servingPort
			servingPort := submitSpec.GetServingPort()
			Expect(servingPort.GetContainer()).To(Equal(int32(1234)))
			Expect(string(servingPort.GetAuthorizationType())).To(Equal(constants.ServingPortAuthorizationTypeAuthorizedUsersOrGroups))
			Expect(servingPort.GetProtocol()).To(Equal(openapi.SERVINGPORTPROTOCOL_HTTP))
			Expect(servingPort.GetAuthorizedUsers()).To(Equal([]string{"user1", "user2"}))
			Expect(servingPort.GetAuthorizedGroups()).To(Equal([]string{"group1", "group2"}))
			Expect(servingPort.GetClusterLocalAccessOnly()).To(Equal(true))

		})

		It("success Inference serving-port syntax sugar", func() {

			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
				"--serving-port=container=1234,authorization-type=authorizedUsersOrGroups,authorized-users=user1:user2,authorized-groups=group1:group2,cluster-local-access-only,protocol=http "+
				"--serving-port=4567",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))

			// servingPort
			// expect the last one to be the one that is set
			servingPort := submitSpec.GetServingPort()
			Expect(servingPort.GetContainer()).To(Equal(int32(4567)))
			Expect(len(servingPort.AuthorizedUsers)).To(Equal(0))
			Expect(len(servingPort.AuthorizedGroups)).To(Equal(0))
			Expect(servingPort.ClusterLocalAccessOnly.Get()).To(BeNil())
			Expect(servingPort.Protocol.Get()).To(BeNil())
			Expect(servingPort.AuthorizationType.Get()).To(BeNil())
		})

		Context("success security", func() {
			It("general security flags", func() {
				output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s --serving-port=%s "+
					"--run-as-uid=100 "+
					"--run-as-gid=200 "+
					"--user-group-source=fromIdpToken "+
					"--seccomp-profile=Localhost "+
					"--supplemental-groups=123",
					InferenceComputeFullName, InferenceComputeFullNameImageName, InferenceServingPort), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
				submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
				Expect(err).To(BeNil())
				Expect(submitSpec.Image.IsSet()).To(Equal(true))
				Expect(*submitSpec.Image.Get()).To(Equal(InferenceComputeFullNameImageName))

				security := submitSpec.GetSecurity()
				Expect(security.GetRunAsUid()).To(Equal(int64(100)))
				Expect(security.GetRunAsGid()).To(Equal(int64(200)))
				Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_FROM_IDP_TOKEN))
				Expect(security.GetSeccompProfileType()).To(Equal(openapi.SECCOMPPROFILETYPE_LOCALHOST))
				Expect(security.GetSupplementalGroups()).To(Equal("123"))
			})

			It("security container user flags", func() {
				output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
					"--run-as-uid=300 "+
					"--run-as-gid=400 "+
					"--seccomp-profile=Localhost "+
					"--supplemental-groups=456 --supplemental-groups=123",
					InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
				submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
				Expect(err).To(BeNil())

				security := submitSpec.GetSecurity()
				Expect(security.GetRunAsUid()).To(Equal(int64(300)))
				Expect(security.GetRunAsGid()).To(Equal(int64(400)))
				Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_CUSTOM))
				Expect(security.GetSeccompProfileType()).To(Equal(openapi.SECCOMPPROFILETYPE_LOCALHOST))
				Expect(security.GetSupplementalGroups()).To(Equal("456,123"))
			})

			Context("run-as-user security flags", func() {
				BeforeEach(func() {
					ResetRunaiCommand()
				})

				It("should pass", func() {
					output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
						"--run-as-user",
						InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
					Expect(err).To(BeNil())
					Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))

					submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
					Expect(err).To(BeNil())
					security := submitSpec.GetSecurity()
					Expect(security.GetRunAsUid()).To(Equal(int64(os.Getuid())))
					Expect(security.GetRunAsGid()).To(Equal(int64(os.Getgid())))
					Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_CUSTOM))

					supGroup, err := os.Getgroups()
					Expect(err).NotTo(HaveOccurred())
					supGroupStr := utils.IntSliceToString(supGroup)
					Expect(security.GetSupplementalGroups()).To(Equal(supGroupStr))
				})
				It("should failed when using mutual exclusive flags - run-as-uid", func() {
					output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
						"--run-as-uid=300 "+
						"--run-as-user",
						InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
				It("should failed when using mutual exclusive flags - run-as-gid", func() {
					output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
						"--run-as-gid=200 "+
						"--run-as-user",
						InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
				It("should failed when using mutual exclusive flags - supplemental-groups", func() {
					output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
						"--supplemental-groups=1,2,3 "+
						"--run-as-user",
						InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
				It("should failed when using mutual exclusive flags - user-group-source", func() {
					output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
						"--user-group-source=fromIdpToken "+
						"--run-as-user",
						InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
			})

			It("capabilities", func() {
				output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
					"--capability SYS_TTY_CONFIG "+
					"--capability LINUX_IMMUTABLE "+
					"--capability NET_BIND_SERVICE "+
					"--capability IPC_OWNER",
					InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
				submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
				Expect(err).To(BeNil())

				security := submitSpec.GetSecurity()
				capabilities := security.GetCapabilities()
				Expect(len(capabilities)).To(Equal(4))
				Expect(capabilities[0]).To(Equal(openapi.CAPABILITY_SYS_TTY_CONFIG))
				Expect(capabilities[1]).To(Equal(openapi.CAPABILITY_LINUX_IMMUTABLE))
				Expect(capabilities[2]).To(Equal(openapi.CAPABILITY_NET_BIND_SERVICE))
				Expect(capabilities[3]).To(Equal(openapi.CAPABILITY_IPC_OWNER))
			})
		})

		It("success scheduling", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s "+
				"--node-pools=node-one --node-pools=node-two "+
				"--toleration=name=NAME,operator=Exists,key=KEY,effect=PreferNoSchedule "+
				"--node-type=the-best-one",
				InferenceComputeFullName, InferenceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferenceComputeFullName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferenceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.GetNodePools()).To(Equal([]string{"node-one", "node-two"}))
			Expect(submitSpec.GetNodeType()).To(Equal("the-best-one"))

			tolerationList := submitSpec.GetTolerations()
			Expect(len(tolerationList)).To(Equal(1))
			Expect(tolerationList[0].GetName()).To(Equal("NAME"))
			Expect(tolerationList[0].GetOperator()).To(Equal(openapi.TOLERATIONOPERATOR_EXISTS))
			Expect(tolerationList[0].GetKey()).To(Equal("KEY"))
			Expect(tolerationList[0].GetEffect()).To(Equal(openapi.TOLERATIONEFFECT_PREFER_NO_SCHEDULE))
		})

		It("prefix prefix - generate inference with prefix name", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit --project=p301 --image=%s --name-prefix=%s",
				InferenceMinimalImageName, InferencePrefixName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferencePrefixName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferencePrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceMinimalImageName))
		})

		It("execute string at the end goes to args", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s -- this is args",
				InferencePrefixName, InferenceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, InferencePrefixName)))
			submitSpec, err := clients.GetSpecByInferenceName(InferencePrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(InferenceMinimalImageName))

			Expect(submitSpec.GetArgs()).To(Equal("this is args"))
		})

	})

	Context("Inference update validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})
		It("inference update no name", func() {
			output, err := cli.RunCommand("inference update --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("workload name is missing, please use 'runai workload <command> <workload-name> [FLAGS]'"))
			Expect(output).To(Equal("Error: an inference update failed. workload name is missing, please use 'runai workload <command> <workload-name> [FLAGS]'\n"))
		})

		It("inference update success - basic flags set", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --image=new-img --image-pull-policy=true --working-dir=/a/b/c --activation-replicas=1 --concurrency-hard-limit=0 --initial-replicas=2 --initialization-timeout-seconds=4 --max-replicas=5 --metric=metric --metric-threshold=6 --metric-threshold-percentage=70 --min-replicas=1 --scale-down-delay-seconds=8 --scale-to-zero-retention-seconds=9", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(updatingMsg, clients.WorkloadName01)))
		})

		It("inference update success - compute fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --cpu-core-request=0.5 --cpu-core-limit=1.0 --cpu-memory-request=1G --cpu-memory-limit=2G --gpu-devices-request=1 --gpu-memory-request=8G --gpu-memory-limit=16G --large-shm", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(updatingMsg, clients.WorkloadName01)))
		})

		It("inference update success - node selection fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --node-pools=pool1 --node-pools=pool2", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(updatingMsg, clients.WorkloadName01)))
		})

		It("inference update success - combined fields", func() {
			output, err := cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --image=new-img --cpu-core-request=0.5 --gpu-devices-request=1 --node-pools=pool1 --min-replicas=1 --max-replicas=5", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(updatingMsg, clients.WorkloadName01)))
		})

	})

})

var _ = Describe("Inference Invalid Test", Ordered, func() {
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
	})

	Context("Pre workload command validation", Ordered, func() {
		It("login validation", func() {
			err = clients.AuthValidation(configPath, cli.ConfigJson, envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})

		It("cluster set", func() {
			err = clients.ClusterSet(envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})
	})

	Context("Inference submit", Ordered, func() {

		It("termination-grace-period doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --termination-grace-period",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --termination-grace-period"))
		})

		It("host-network flag is recognized", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s --host-network",
				InferenceMinimalName, InferenceMinimalImageName), envMap)
			if err != nil {
				Expect(err.Error()).ToNot(Equal("unknown flag: --host-network"))
			}
		})

		It("host-ipc flag is recognized", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --image=%s --host-ipc",
				InferenceMinimalName, InferenceMinimalImageName), envMap)
			if err != nil {
				Expect(err.Error()).ToNot(Equal("unknown flag: --host-ipc"))
			}
		})

		It("allow-privilege-escalation doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --allow-privilege-escalation",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --allow-privilege-escalation"))
		})

		It("backoff-limit doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --backoff-limit",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --backoff-limit"))
		})

		It("auto-deletion-time-after-completion doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --auto-deletion-time-after-completion",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --auto-deletion-time-after-completion"))
		})

		It("s3 doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --s3",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --s3"))
		})

		It("completions doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --completions",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --completions"))
		})

		It("parallelism doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --parallelism",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --parallelism"))
		})

		It("cannot have args and command together ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference submit %s --project=p301 --name-prefix=shouldnotwork",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("both prefix and name are set, only one can be used"))
		})
	})

	Context("Inference update", Ordered, func() {

		It("cannot have name-prefix flag", func() {
			_, err = cli.RunCommand(fmt.Sprintf("inference update %s --project=p301 --name-prefix=shouldnotwork",
				InferenceMinimalName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --name-prefix"))
		})

	})
})

func ResetRunaiCommand() {
	// Reset the inference command by removing all subcommands
	cmd := inference.Cmd
	for _, c := range cmd.Commands() {
		cmd.RemoveCommand(c)
	}

	// Re-initialize the inference command structure with both standard and distributed
	inference.InitInferenceCmd()
}
