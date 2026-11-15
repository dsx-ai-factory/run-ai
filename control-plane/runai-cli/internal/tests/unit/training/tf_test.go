/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package training

import (
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"

	cmdfactory "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/training/tensorflow"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
)

const (
	tensorflowTestFilePath = "tensorflow_training"
)

const (
	TensorFlowMinimalName      = "tr-minimal"
	TensorFlowMinimalImageName = "minimal-image"

	TensorFlowComputeFullName      = "tr-compute-full"
	TensorFlowComputeFullImageName = "compute-full-image"
	TensorFlowPrefixName           = "tensorflow-prefix"
)

var submitTfList = []string{TensorFlowMinimalName, TensorFlowComputeFullName, TensorFlowPrefixName}

var _ = Describe("TensorFlow Training Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, tensorflowTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())

		// Initialize the token manager with a fake backend always returning a valid token
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)

		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Distributed{Framework: api.TF}, submitTfList)
		resetTensorFlowCommand()
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
	})

	Context("TensorFlow list validation", Ordered, func() {
		BeforeEach(func() {
			resetTensorFlowCommand()
		})
		It("TensorFlow list all", func() {
			output, err := cli.RunCommand("training tensorflow list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
		})

		It("TensorFlow shortcut sanity", func() {
			output, err := cli.RunCommand("tensorflow list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
		})

		It("TensorFlow list all and project flags", func() {
			_, err := cli.RunCommand("training tensorflow list -A --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("[all project] are set none of the others can be;"))
		})

		It("TensorFlow list project p301", func() {
			output, err := cli.RunCommand("training tensorflow list --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))
		})

		It("TensorFlow list project p300", func() {
			output, err := cli.RunCommand("training tensorflow list --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedListWithProjectThatDoesNotExist))
		})
	})

	Context("TensorFlow delete validation", Ordered, func() {
		BeforeEach(func() {
			resetTensorFlowCommand()
		})
		It("TensorFlow delete no uuid or name", func() {
			output, err := cli.RunCommand("training tensorflow delete --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to delete workload. one of the attributes must be set name|uuid\n"))
		})

		It("TensorFlow delete, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow delete --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf("Deleting workload '%s'...\n", clients.UuidExist)))
		})

		It("TensorFlow delete, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow delete --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to delete workload. generic error string\n"))
		})
	})
	Context("TensorFlow submit validation", Ordered, func() {
		BeforeEach(func() {
			resetTensorFlowCommand()
		})
		It("TensorFlow submit - no project", func() {
			output, err := cli.RunCommand("training tensorflow submit", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix(MissingProject))
		})

		It("TensorFlow submit - missing flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedListWithProjectThatDoesNotExist))
		})

		It("TensorFlow submit - project does not exist", func() {
			output, err := cli.RunCommand("training tensorflow submit ws01 --project=p300 --image=test-image", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedListWithProjectThatDoesNotExist))
		})

		It("TensorFlow submit - success minimal", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 4",
				TensorFlowMinimalName, TensorFlowMinimalImageName), envMap)
			Expect(err).To(BeNil())
			outputStr := fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowMinimalName)
			Expect(output).To(HavePrefix(outputStr))
			Expect(output).To(Equal(fmt.Sprintf("%s\nTo track the workload's status, run 'runai training tf describe %s'\n", outputStr, TensorFlowMinimalName)))

		})

		It("TensorFlow submit - success with compute resources", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 3 "+
				"--gpu-devices-request 1",
				TensorFlowComputeFullName, TensorFlowComputeFullImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowComputeFullName)))

		})

		It("TensorFlow submit - with prefix", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit --project=p301 --image=%s --name-prefix=%s "+
				"--workers 2",
				TensorFlowMinimalImageName, TensorFlowPrefixName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowPrefixName)))

		})
		It("TensorFlow submit - success with container configuration", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 3 "+
				"--image-pull-policy=IfNotPresent "+
				"--working-dir=/test/path "+
				"--create-home-dir "+
				"--environment-variable=env1=value1 "+
				"--preferred-pod-topology-key=IS_PREF "+
				"-- args should have the value",
				TensorFlowComputeFullName, TensorFlowComputeFullImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowComputeFullName)))

			submitSpec, err := clients.GetSpecByDistributedName(TensorFlowComputeFullName)
			Expect(err).To(BeNil())

			workerSpec := submitSpec.GetSpec()
			Expect(workerSpec.Image.IsSet()).To(Equal(true))
			Expect(*workerSpec.Image.Get()).To(Equal(TensorFlowComputeFullImageName))
			Expect(workerSpec.GetNumWorkers()).To(Equal(int32(3)))

			Expect(workerSpec.GetImagePullPolicy()).To(Equal(openapi.IMAGEPULLPOLICY_IF_NOT_PRESENT))
			Expect(workerSpec.GetWorkingDir()).To(Equal("/test/path"))
			Expect(workerSpec.GetCreateHomeDir()).To(Equal(true))
			Expect(workerSpec.GetArgs()).To(Equal("args should have the value"))

			pa := workerSpec.GetPodAffinity()
			Expect(pa.GetType()).To(Equal(openapi.PODAFFINITYTYPE_PREFERRED))
			Expect(pa.GetKey()).To(Equal("IS_PREF"))

			env := workerSpec.GetEnvironmentVariables()
			Expect(len(env)).To(Equal(1))
			Expect(env[0].GetName()).To(Equal("env1"))
			Expect(env[0].GetValue()).To(Equal("value1"))
		})

		It("TensorFlow submit - success with compute resources", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 2 "+
				"--cpu-memory-request=2G --cpu-memory-limit=1G "+
				"--cpu-core-limit=0.2 --cpu-core-request=0.3 "+
				"--gpu-devices-request=2 "+
				"--gpu-memory-limit=20M --gpu-memory-request=30M "+
				"--gpu-portion-limit=0.23 --gpu-portion-request=0.43 "+
				"--gpu-request-type=memory "+
				"--large-shm=true",
				TensorFlowComputeFullName, TensorFlowComputeFullImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowComputeFullName)))

			submitSpec, err := clients.GetSpecByDistributedName(TensorFlowComputeFullName)
			Expect(err).To(BeNil())

			workerSpec := submitSpec.GetSpec()
			compute := workerSpec.Compute.Get()
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
		It("tf submit - success with master CPU flags", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 2 "+
				"--cpu-memory-request=0.5G --cpu-memory-limit=1G "+
				"--cpu-core-limit=0.9 --cpu-core-request=0.1 "+
				"--master-cpu-core-limit 0.5 "+
				"--master-cpu-core-request 0.2 "+
				"--master-cpu-memory-limit 1G "+
				"--master-cpu-memory-request=500M",
				TensorFlowComputeFullName, TensorFlowComputeFullImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowComputeFullName)))

			submitSpec, err := clients.GetSpecByDistributedName(TensorFlowComputeFullName)
			Expect(err).To(BeNil())

			masterSpec := submitSpec.GetMasterSpec()
			compute := masterSpec.Compute.Get()
			Expect(*compute.CpuCoreLimit.Get()).To(Equal(0.5))
			Expect(*compute.CpuCoreRequest.Get()).To(Equal(0.2))
			Expect(*compute.CpuMemoryLimit.Get()).To(Equal("1G"))
			Expect(*compute.CpuMemoryRequest.Get()).To(Equal("500M"))

			workerSpec := submitSpec.GetSpec()
			compute = workerSpec.Compute.Get()
			Expect(*compute.CpuCoreLimit.Get()).To(Equal(0.9))
			Expect(*compute.CpuCoreRequest.Get()).To(Equal(0.1))
			Expect(*compute.CpuMemoryLimit.Get()).To(Equal("1G"))
			Expect(*compute.CpuMemoryRequest.Get()).To(Equal("0.5G"))
		})
		It("TensorFlow submit - success with storage configuration", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 2 "+
				"--configmap-map-volume=name=CONFIGMAP_NAME,path=PATH "+
				"--existing-pvc=claimname=CLAIM_NAME_EXIST,path=PATH_EXIST_PVC "+
				"--new-pvc=claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral "+
				"--git-sync=name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION "+
				"--host-path=path=PATH,mount=MOUNT,mount-propagation=HostToContainer,readwrite "+
				"--nfs=path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite "+
				"--s3=name=NAME,bucket=BUCKET,path=PATH,accesskey=ACCESS_KEY,url=URL",
				TensorFlowComputeFullName, TensorFlowComputeFullImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowComputeFullName)))

			submitSpec, err := clients.GetSpecByDistributedName(TensorFlowComputeFullName)
			Expect(err).To(BeNil())

			workerSpec := submitSpec.GetSpec()
			storage := workerSpec.Storage.Get()

			Expect(len(storage.HostPath)).To(Equal(1))
			Expect(storage.HostPath[0].GetPath()).To(Equal("PATH"))
			Expect(storage.HostPath[0].GetMountPath()).To(Equal("MOUNT"))
			Expect(storage.HostPath[0].GetReadOnly()).To(Equal(false))

			Expect(len(storage.Nfs)).To(Equal(1))
			Expect(storage.Nfs[0].GetPath()).To(Equal("PATH"))
			Expect(storage.Nfs[0].GetServer()).To(Equal("SERVER"))
			Expect(storage.Nfs[0].GetMountPath()).To(Equal("MOUNT_PATH"))
			Expect(storage.Nfs[0].GetReadOnly()).To(Equal(false))

			Expect(len(storage.Pvc)).To(Equal(2))
			Expect(storage.Pvc[0].GetExistingPvc()).To(Equal(false))
			Expect(storage.Pvc[0].GetClaimName()).To(Equal("CLAIM_NAME"))
			Expect(storage.Pvc[0].GetPath()).To(Equal("PATH"))
			Expect(storage.Pvc[0].GetReadOnly()).To(Equal(true))
			Expect(storage.Pvc[0].GetEphemeral()).To(Equal(true))

			Expect(len(storage.Git)).To(Equal(1))
			Expect(storage.Git[0].GetName()).To(Equal("NAME"))
			Expect(storage.Git[0].GetRepository()).To(Equal("REPO"))
			Expect(storage.Git[0].GetPath()).To(Equal("PATH"))
			Expect(storage.Git[0].GetPasswordSecret()).To(Equal("SECRET"))
			Expect(storage.Git[0].GetRevision()).To(Equal("REVISION"))

			Expect(len(storage.S3)).To(Equal(1))
			Expect(storage.S3[0].GetName()).To(Equal("NAME"))
			Expect(storage.S3[0].GetBucket()).To(Equal("BUCKET"))
			Expect(storage.S3[0].GetPath()).To(Equal("PATH"))
			Expect(storage.S3[0].GetAccessKeySecret()).To(Equal("ACCESS_KEY"))
			Expect(storage.S3[0].GetUrl()).To(Equal("URL"))
		})

		It("TensorFlow submit - success with network configuration", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 2 "+
				"--port=service-type=NodePort,container=80,external=8080 "+
				"--external-url=container=9443,url=https://external.runai.com,authusers=user1:user2,authgroups=group1:group3",
				TensorFlowComputeFullName, TensorFlowComputeFullImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowComputeFullName)))

			submitSpec, err := clients.GetSpecByDistributedName(TensorFlowComputeFullName)
			Expect(err).To(BeNil())

			workerSpec := submitSpec.GetSpec()
			ports := workerSpec.GetPorts()
			Expect(len(ports)).To(Equal(1))
			Expect(ports[0].GetContainer()).To(Equal(int32(80)))
			Expect(ports[0].GetExternal()).To(Equal(int32(8080)))
			Expect(ports[0].GetServiceType()).To(Equal(openapi.PORTSERVICETYPE_NODE_PORT))

			exposedUrls := workerSpec.GetExposedUrls()
			Expect(len(exposedUrls)).To(Equal(1))
			Expect(exposedUrls[0].GetContainer()).To(Equal(int32(9443)))
			Expect(exposedUrls[0].GetUrl()).To(Equal("https://external.runai.com"))
			Expect(exposedUrls[0].GetAuthorizedUsers()).To(Equal([]string{"user1", "user2"}))
			Expect(exposedUrls[0].GetAuthorizedGroups()).To(Equal([]string{"group1", "group3"}))
		})

		It("TensorFlow submit - success with security configuration", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training tensorflow submit %s --project=p301 --image=%s "+
				"--workers 2 "+
				"--allow-privilege-escalation "+
				"--run-as-uid=100 "+
				"--run-as-gid=200 "+
				"--user-group-source=fromIdpToken "+
				"--seccomp-profile=Localhost "+
				"--supplemental-groups=123 "+
				"--host-ipc "+
				"--host-network "+
				"--capability SYS_TTY_CONFIG --capability LINUX_IMMUTABLE",
				TensorFlowComputeFullName, TensorFlowComputeFullImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectDistSuccessMsg, TensorFlowComputeFullName)))

			submitSpec, err := clients.GetSpecByDistributedName(TensorFlowComputeFullName)
			Expect(err).To(BeNil())

			workerSpec := submitSpec.GetSpec()
			security := workerSpec.GetSecurity()
			Expect(security.GetAllowPrivilegeEscalation()).To(Equal(true))
			Expect(security.GetRunAsUid()).To(Equal(int64(100)))
			Expect(security.GetRunAsGid()).To(Equal(int64(200)))
			Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_FROM_IDP_TOKEN))
			Expect(security.GetSeccompProfileType()).To(Equal(openapi.SECCOMPPROFILETYPE_LOCALHOST))
			Expect(security.GetSupplementalGroups()).To(Equal("123"))
			Expect(security.GetHostIpc()).To(Equal(true))
			Expect(security.GetHostNetwork()).To(Equal(true))

			capabilities := security.GetCapabilities()
			Expect(len(capabilities)).To(Equal(2))
			Expect(capabilities).To(ContainElement(openapi.CAPABILITY_SYS_TTY_CONFIG))
			Expect(capabilities).To(ContainElement(openapi.CAPABILITY_LINUX_IMMUTABLE))
		})

	})
	Context("TensorFlow submit flags", Ordered, func() {
		BeforeEach(func() {
			resetTensorFlowCommand()
		})
		It("should not show master gpu flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())
			Expect(output).ToNot(ContainSubstring(SlotPerWorkerFlag))
			Expect(output).ToNot(ContainSubstring(SegmentSize))
		})

		It("should show all expected general flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			generalFlags := []string{Annotation, LabelFlag, NamePrefix, Project}
			for _, flag := range generalFlags {
				Expect(output).To(ContainSubstring(flag), "General flag %s should be present in help output", flag)
			}
		})

		It("should show all expected container flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			containerFlags := []string{Command, CreateHomeDir, Environment, Image, ImagePullPolicy, MasterArgs, MasterEnvironment, PodTopologyKey, RequiredTopologyKey, Stdin, Tty, WorkingDir}
			for _, flag := range containerFlags {
				Expect(output).To(ContainSubstring(flag), "Container flag %s should be present in help output", flag)
			}
		})

		It("should show all expected compute flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			computeFlags := []string{CpuCoreLimit, CpuCoreRequest, CpuMemoryLimit, CpuMemoryRequest, ExtendedResource, GpuDevicesRequest, GpuMemoryLimit, GpuMemoryRequest,
				GpuMemoryPortionLimit, GpuMemoryPortionRequest, GpuComputeMode,
				GpuRequestType, LargeShm, MasterExtendedResource}
			for _, flag := range computeFlags {
				Expect(output).To(ContainSubstring(flag), "Compute flag %s should be present in help output", flag)
			}
		})

		It("should show all expected storage flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			storageFlags := []string{ConfigMapVolume, ExistingPvc, GitSync, HostPath, MasterNoPvcs, NewPvc, Nfs, S3}
			for _, flag := range storageFlags {
				Expect(output).To(ContainSubstring(flag), "Storage flag %s should be present in help output", flag)
			}
		})

		It("should show all expected network flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			networkFlags := []string{ExternalUrl, Port}
			for _, flag := range networkFlags {
				Expect(output).To(ContainSubstring(flag), "Network flag %s should be present in help output", flag)
			}
		})

		It("should show all expected lifecycle flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			lifecycleFlags := []string{AutoDeletionTime, BackoffLimit}
			for _, flag := range lifecycleFlags {
				Expect(output).To(ContainSubstring(flag), "Lifecycle flag %s should be present in help output", flag)
			}
		})

		It("should show all expected security flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			securityFlags := []string{AllowPrivilegeEscalation, Capability, HostIpc, HostNetwork, RunAsGid, RunAsUid,
				RunAsUser, SeccompProfile, SupplementalGroups, UserGroupSource}
			for _, flag := range securityFlags {
				Expect(output).To(ContainSubstring(flag), "Security flag %s should be present in help output", flag)
			}
		})

		It("should show all expected scheduling flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			schedulingFlags := []string{NodePools, NodeType, Toleration}
			for _, flag := range schedulingFlags {
				Expect(output).To(ContainSubstring(flag), "Scheduling flag %s should be present in help output", flag)
			}
		})

		It("should show all expected distributed flags", func() {
			output, err := cli.RunCommand("training tensorflow submit --help", envMap)
			Expect(err).To(BeNil())

			distributedFlags := []string{Workers}
			for _, flag := range distributedFlags {
				Expect(output).To(ContainSubstring(flag), "Distributed flag %s should be present in help output", flag)
			}
		})
	})
	DescribeTable("TensorFlow workload name argument auto-completion", Ordered,
		func(action string) {
			resetMpiCommand()
			output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(fmt.Sprintf("training tensorflow %s --project=p301", action), ""), envMap)
			Expect(err).To(BeNil())
			factories.AssertCompletionOptions(output, []string{clients.WorkloadName01})
			factories.AssertCompletionResponseStatus(output, false)
		},
		Entry("describe", "describe"),
		Entry("exec", "exec"),
		Entry("attach", "attach"),
		Entry("delete", "delete"),
		Entry("logs", "logs"),
		Entry("port-forward", "port-forward"),
	)
})

func resetTensorFlowCommand() {
	cmd := tensorflow.Cmd
	for _, c := range cmd.Commands() {
		cmd.RemoveCommand(c)
	}
	cmdfactory.WorkloadsCmdInit(cmd, api.TF, api.Distributed)
}
