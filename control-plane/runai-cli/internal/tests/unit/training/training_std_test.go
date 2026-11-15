/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package training

import (
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/training/runaijob"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	trainingStdTestFilePath = "std_training"

	ExpectedTrainingListWithProjectThatDoesNotExist = "project name: p300, not found\n"
	ExpectedTrainingMissingFlags                    = "required flag(s) \"image\" not set"
	ExpectSuccessMsg                                = "Creating training %s..."
)

const (
	TrainingMinimalName      = "tr-minimal"
	TrainingMinimalImageName = "minimal-image"

	TrainingComputeFullName          = "tr-compute-full"
	TrainingComputeFullNameImageName = "compute-full-image"
	TrainingPrefixName               = "training-prefix"
)

var submitTrainingList = []string{TrainingMinimalName, TrainingComputeFullName, TrainingPrefixName}

var _ = Describe("Training Standard Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, trainingStdTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusterStatus(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Training{}, submitTrainingList)
		ResetRunaiCommand()
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

	Context("Training standard list validation", Ordered, func() {
		It("Training standard list all", func() {
			output, err := cli.RunCommand("training standard list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName01))
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
		})

		It("Training standard with framework - no such flag", func() {
			_, err := cli.RunCommand("training standard list --framework runai", envMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unknown flag: --framework"))
		})

		It("Training list all and project flags", func() {
			_, err := cli.RunCommand("training standard list -A --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("[all project] are set none of the others can be;"))
		})

		It("Training list project p301", func() {
			output, err := cli.RunCommand("training standard list --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName01))
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))
		})

		It("Training list project p300", func() {
			output, err := cli.RunCommand("training standard list --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedTrainingListWithProjectThatDoesNotExist))
		})

		Describe("Training list project flags", func() {
			BeforeAll(func() {
				err := clients.ProjectSet(clients.ProjectOneName, envMap)
				if err != nil {
					Fail("project set failed, error: " + err.Error())
				}
			})

			AfterAll(func() {
				// workarounds for the fact that the project is not cleared after each test
				err := clients.ClusterSet(envMap)
				if err != nil {
					Fail("cluster set failed, error: " + err.Error())
				}
			})

			It("Training list project with default project", func() {
				output, err := cli.RunCommand("training standard list", envMap)
				Expect(err).To(BeNil())
				Expect(output).ToNot(ContainSubstring(clients.WorkloadName01))
				Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))

			})

			It("Training list all projects with default project set", func() {
				output, err := cli.RunCommand("training standard list --all", envMap)
				Expect(err).To(BeNil())
				Expect(output).ToNot(ContainSubstring(clients.WorkloadName01))
				Expect(output).To(ContainSubstring(clients.WorkloadName02))
			})
		})
	})

	Context("Training delete validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("Training delete no uuid or name", func() {
			output, err := cli.RunCommand("training standard delete --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to delete workload. one of the attributes must be set name|uuid\n"))
		})

		It("training standard delete, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard delete --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf("Deleting workload '%s'...\n", clients.UuidExist)))
		})

		It("training standard delete, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard delete --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to delete workload. generic error string\n"))
		})
	})

	Context("Training suspend validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("Training suspend no uuid or name", func() {
			output, err := cli.RunCommand("training standard suspend --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to suspend workload. one of the attributes must be set name|uuid\n"))
		})

		It("Training suspend both uuid and name", func() {
			output, err := cli.RunCommand("training standard suspend anyname --project=p301 --uuid=anothername", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to suspend workload. one of the attributes must be set name|uuid\n"))
		})

		It("Training suspend, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard suspend --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf("Stopping workload '%v'...\nTo track the workload's status, run 'runai training standard describe %v'\n", clients.UuidExist, clients.UuidExist)))
		})

		It("Training suspend, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard suspend --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to suspend workload. generic error string\n"))
		})

		It("Training suspend, name - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard suspend %s --project=p302", clients.WorkloadName021), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf("Stopping workload '%v'...\nTo track the workload's status, run 'runai training standard describe %v'", clients.WorkloadName021, clients.WorkloadName021)))
		})

		It("Training suspend, name - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard suspend %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix("Error: failed to suspend workload. generic error string\n"))
		})
	})

	Context("Training resume validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("training standard resume no uuid or name", func() {
			output, err := cli.RunCommand("training standard resume --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to resume workload. one of the attributes must be set name|uuid\n"))
		})

		It("training standard resume both uuid and name", func() {
			output, err := cli.RunCommand("training standard resume anyname --project=p301 --uuid=anothername", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to resume workload. one of the attributes must be set name|uuid\n"))
		})

		It("training standard resume, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard resume --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf("Resuming workload '%v'...\nTo track the workload's status, run 'runai training standard describe %v'\n", clients.UuidExist, clients.UuidExist)))
		})

		It("training standard resume, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard resume --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to resume workload. generic error string\n"))
		})

		It("Training resume, name - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard resume %s --project=p302", clients.WorkloadName021), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf("Resuming workload '%v'...\nTo track the workload's status, run 'runai training standard describe %v'\n", clients.WorkloadName021, clients.WorkloadName021)))
		})

		It("Training resume, name - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard resume %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix("Error: failed to resume workload. generic error string\n"))
		})
	})

	Context("Training submit prerequisite validation", Ordered, func() {
		It("training standard submit - no project", func() {
			output, err := cli.RunCommand("training standard submit", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix(MissingProject))
		})

		It("training standard submit - missing flags", func() {
			output, err := cli.RunCommand("training standard submit --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedTrainingListWithProjectThatDoesNotExist))
		})

		It("training standard submit - project does not exist", func() {
			output, err := cli.RunCommand("training standard submit ws01 --project=p300 --image=lies", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedTrainingListWithProjectThatDoesNotExist))
		})
	})

	Context("Training submit validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})
		It("training standard submit - success naive", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s",
				TrainingMinimalName, TrainingMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingMinimalName)))
			submitSpec, err := clients.GetSpecByTrainingName(TrainingMinimalName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingMinimalImageName))

			Expect(submitSpec.GetCreateHomeDir()).To(Equal(false))
		})

		It("Training submit - success container", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--image-pull-policy=IfNotPresent "+
				"--tty --stdin "+
				"--working-dir=/test/path "+
				"--create-home-dir "+
				"--environment-variable=env1=value1 "+
				"--preferred-pod-topology-key=IS_PREF "+
				"-- args should have the value",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))

			Expect(submitSpec.GetCommand()).To(Equal(""))
			Expect(submitSpec.GetImagePullPolicy()).To(Equal(openapi.IMAGEPULLPOLICY_IF_NOT_PRESENT))
			Expect(submitSpec.GetWorkingDir()).To(Equal("/test/path"))
			Expect(submitSpec.GetTty()).To(Equal(true))
			Expect(submitSpec.GetStdin()).To(Equal(true))
			Expect(submitSpec.GetCreateHomeDir()).To(Equal(true))
			Expect(submitSpec.GetArgs()).To(Equal("args should have the value"))

			pa := submitSpec.GetPodAffinity()
			Expect(pa.GetType()).To(Equal(openapi.PODAFFINITYTYPE_PREFERRED))
			Expect(pa.GetKey()).To(Equal("IS_PREF"))

			env := submitSpec.GetEnvironmentVariables()
			Expect(len(env)).To(Equal(1))
			Expect(env[0].GetName()).To(Equal("env1"))
			Expect(env[0].GetValue()).To(Equal("value1"))
		})

		It("training standard submit - success compute", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--cpu-memory-request=2G --cpu-memory-limit=1G "+
				"--cpu-core-limit=0.2 --cpu-core-request=0.3 "+
				"--gpu-devices-request=2 "+
				"--gpu-memory-limit=20M --gpu-memory-request=30M "+
				"--gpu-portion-limit=0.23 --gpu-portion-request=0.43 "+
				"--gpu-request-type=memory "+
				"--large-shm=true",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))
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

		It("training standard submit - success general", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--annotation=ankey1=anvalue1 --annotation=ankey2=anvalue2 "+
				"--label=lkey1=lvalue1 --label=lkey2=lvalue2",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))
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

		It("training standard submit - success storage", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--configmap-map-volume=name=CONFIGMAP_NAME,path=PATH "+
				"--existing-pvc=claimname=CLAIM_NAME_EXIST,path=PATH_EXIST_PVC "+
				"--new-pvc=claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral "+
				"--git-sync=name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION "+
				"--host-path=path=PATH,mount=MOUNT,mount-propagation=HostToContainer,readwrite "+
				"--nfs=path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite "+
				"--s3=name=NAME,bucket=BUCKET,path=PATH,accesskey=ACCESS_KEY,url=URL",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))
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

			Expect(len(storage.S3)).To(Equal(1))
			// name=NAME,bucket=BUCKET,path=PATH,accesskey=ACCESS_KEY,url=URL
			Expect(storage.S3[0].GetName()).To(Equal("NAME"))
			Expect(storage.S3[0].GetBucket()).To(Equal("BUCKET"))
			Expect(storage.S3[0].GetPath()).To(Equal("PATH"))
			Expect(storage.S3[0].GetAccessKeySecret()).To(Equal("ACCESS_KEY"))
			Expect(storage.S3[0].GetUrl()).To(Equal("URL"))

			Expect(len(storage.ConfigMapVolume)).To(Equal(1))
			// name=CONFIGMAP_NAME,path=PATH
			Expect(storage.ConfigMapVolume[0].GetConfigMap()).To(Equal("CONFIGMAP_NAME"))
			Expect(storage.ConfigMapVolume[0].GetMountPath()).To(Equal("PATH"))
		})

		It("training standard submit - success network", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--port=service-type=NodePort,container=80,external=8080 "+
				"--external-url=container=9443,url=https://external.runai.com,authusers=user1:user2,authgroups=group1:group3",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))
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

		It("training standard submit - success lifecycle", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--auto-deletion-time-after-completion=5h "+
				"--termination-grace-period 5m "+
				"--backoff-limit=9 "+
				"--runs=11 "+
				"--parallelism=12",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))

			Expect(*submitSpec.AutoDeletionTimeAfterCompletionSeconds.Get()).To(Equal(int64(18000)))
			Expect(*submitSpec.TerminationGracePeriodSeconds.Get()).To(Equal(int64(300)))
			Expect(*submitSpec.BackoffLimit.Get()).To(Equal(int64(9)))
			Expect(submitSpec.HasPriorityClass()).To(Equal(false))
			Expect(*submitSpec.Completions.Get()).To(Equal(int32(11)))
			Expect(*submitSpec.Parallelism.Get()).To(Equal(int32(12)))
		})

		It("training standard submit - success lifecycle - no grace period generated (we want the k8s default)", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--auto-deletion-time-after-completion=5h "+
				"--backoff-limit=9 "+
				"--runs=11 "+
				"--parallelism=12",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))

			Expect(*submitSpec.AutoDeletionTimeAfterCompletionSeconds.Get()).To(Equal(int64(18000)))
			Expect(submitSpec.TerminationGracePeriodSeconds.Get()).To(BeNil())
			Expect(*submitSpec.BackoffLimit.Get()).To(Equal(int64(9)))
			Expect(submitSpec.HasPriorityClass()).To(Equal(false))
			Expect(*submitSpec.Completions.Get()).To(Equal(int32(11)))
			Expect(*submitSpec.Parallelism.Get()).To(Equal(int32(12)))
		})

		It("training standard submit - success lifecycle", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--auto-deletion-time-after-completion=5h "+
				"--termination-grace-period= "+
				"--backoff-limit=9 "+
				"--runs=11 "+
				"--parallelism=12",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("flag: time: invalid duration"))
			Expect(output).To(Equal(""))
		})

		It("training standard submit - deprecated message on completions", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--auto-deletion-time-after-completion=5h "+
				"--backoff-limit=9 "+
				"--completions=11 "+
				"--parallelism=12",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix("Flag --completions has been deprecated, Use --runs instead"))
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))

			Expect(*submitSpec.AutoDeletionTimeAfterCompletionSeconds.Get()).To(Equal(int64(18000)))
			Expect(*submitSpec.BackoffLimit.Get()).To(Equal(int64(9)))
			Expect(submitSpec.HasPriorityClass()).To(Equal(false))
			Expect(*submitSpec.Completions.Get()).To(Equal(int32(11)))
			Expect(*submitSpec.Parallelism.Get()).To(Equal(int32(12)))
		})

		It("training standard submit - success security", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--allow-privilege-escalation "+
				"--run-as-uid=100 "+
				"--run-as-gid=200 "+
				"--seccomp-profile=Localhost "+
				"--supplemental-groups=123 "+
				"--host-ipc "+
				"--host-network "+
				"--privileged",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))

			security := submitSpec.GetSecurity()
			Expect(security.GetAllowPrivilegeEscalation()).To(Equal(true))
			Expect(security.GetRunAsUid()).To(Equal(int64(100)))
			Expect(security.GetRunAsGid()).To(Equal(int64(200)))
			Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_CUSTOM))
			Expect(security.GetSeccompProfileType()).To(Equal(openapi.SECCOMPPROFILETYPE_LOCALHOST))
			Expect(security.GetSupplementalGroups()).To(Equal("123"))
			Expect(security.GetHostIpc()).To(Equal(true))
			Expect(security.GetHostNetwork()).To(Equal(true))
			Expect(security.GetPrivileged()).To(Equal(true))

			output, err = cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--run-as-uid=300 "+
				"--run-as-gid=400 "+
				"--seccomp-profile=Localhost "+
				"--supplemental-groups=456",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err = clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())

			security = submitSpec.GetSecurity()
			Expect(security.GetAllowPrivilegeEscalation()).To(Equal(false))
			Expect(security.GetRunAsUid()).To(Equal(int64(300)))
			Expect(security.GetRunAsGid()).To(Equal(int64(400)))
			Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_CUSTOM))
			Expect(security.GetSeccompProfileType()).To(Equal(openapi.SECCOMPPROFILETYPE_LOCALHOST))
			Expect(security.GetSupplementalGroups()).To(Equal("456"))
			Expect(security.GetHostIpc()).To(Equal(false))
			Expect(security.GetHostNetwork()).To(Equal(false))
			Expect(security.GetPrivileged()).To(Equal(false))

			output, err = cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--capability SYS_TTY_CONFIG "+
				"--capability LINUX_IMMUTABLE "+
				"--capability NET_BIND_SERVICE "+
				"--capability IPC_OWNER",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err = clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())

			security = submitSpec.GetSecurity()
			capabilities := security.GetCapabilities()
			Expect(len(capabilities)).To(Equal(4))
			Expect(capabilities[0]).To(Equal(openapi.CAPABILITY_SYS_TTY_CONFIG))
			Expect(capabilities[1]).To(Equal(openapi.CAPABILITY_LINUX_IMMUTABLE))
			Expect(capabilities[2]).To(Equal(openapi.CAPABILITY_NET_BIND_SERVICE))
			Expect(capabilities[3]).To(Equal(openapi.CAPABILITY_IPC_OWNER))
		})

		It("training standard submit - success scheduling", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
				"--node-pools=node-one --node-pools=node-two "+
				"--toleration=name=NAME,operator=Exists,key=KEY,effect=PreferNoSchedule "+
				"--node-type=the-best-one "+
				"--priority=build",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.GetNodePools()).To(Equal([]string{"node-one", "node-two"}))
			Expect(submitSpec.GetNodeType()).To(Equal("the-best-one"))

			tolerationList := submitSpec.GetTolerations()
			Expect(len(tolerationList)).To(Equal(1))
			Expect(tolerationList[0].GetName()).To(Equal("NAME"))
			Expect(tolerationList[0].GetOperator()).To(Equal(openapi.TOLERATIONOPERATOR_EXISTS))
			Expect(tolerationList[0].GetKey()).To(Equal("KEY"))
			Expect(tolerationList[0].GetEffect()).To(Equal(openapi.TOLERATIONEFFECT_PREFER_NO_SCHEDULE))

			// validate priority class
			Expect(*submitSpec.PriorityClass.Get()).To(Equal("build"))
		})

		It("training standard submit - prefix prefix - generate training with prefix name", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit --project=p301 --image=%s --name-prefix=%s",
				TrainingMinimalImageName, TrainingPrefixName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingPrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingPrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingMinimalImageName))
		})

		It("training standard submit - execute string at the end goes to args", func() {
			output, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s -- this is args",
				TrainingPrefixName, TrainingMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, TrainingPrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingPrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingMinimalImageName))

			Expect(submitSpec.GetArgs()).To(Equal("this is args"))
		})

		Context("Training submit validation - Attach", Ordered, func() {
			It("training submit - attach with cluster health check error - fail", func() {
				_, err := cli.RunCommand(fmt.Sprintf("training submit -p=p301 --image=%s --attach",
					TrainingMinimalImageName), envMap)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("bla"))
			})

			It("training submit - attach with cluster health check error - fail", func() {
				_, err := cli.RunCommand(fmt.Sprintf("training submit %s -p=p301 --image=%s --attach --wait-for-submit=1s",
					TrainingPrefixName, TrainingMinimalImageName), envMap)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("workload's waiting time was exceeded"))
			})
		})
	})

	DescribeTable("Training workload name argument auto-completion", Ordered,
		func(action string) {
			resetMpiCommand()
			output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(fmt.Sprintf("training %s --project=p302", action), ""), envMap)
			Expect(err).To(BeNil())
			factories.AssertCompletionOptions(output, []string{clients.WorkloadName02})
			factories.AssertCompletionResponseStatus(output, false)
		},
		Entry("describe", "describe"),
		Entry("exec", "exec"),
		Entry("attach", "attach"),
		Entry("delete", "delete"),
		Entry("logs", "logs"),
		Entry("port-forward", "port-forward"),
		Entry("suspend", "suspend"),
		Entry("resume", "resume"),
	)
})

var _ = Describe("Training Invalid Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, trainingTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())

		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Training{}, submitTrainingList)
	})

	Context("Pre workload command validation", Ordered, func() {
		It("login validation", func() {
			err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})

		It("cluster set", func() {
			err := clients.ClusterSet(envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})
	})

	It("training standard submit - preemptible doesn't exist ", func() {
		_, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s "+
			"--preemptible",
			TrainingMinimalName, TrainingMinimalImageName), envMap)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("unknown flag: --preemptible"))
	})

	It("workspace submit - cannot have args and command together ", func() {
		_, err := cli.RunCommand(fmt.Sprintf("training standard submit %s --project=p301 --image=%s --name-prefix=shouldnotwork",
			TrainingMinimalName, TrainingMinimalImageName), envMap)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("both prefix and name are set, only one can be used"))
	})
})

func ResetRunaiCommand() {
	cmd := runaijob.Cmd
	for _, c := range cmd.Commands() {
		cmd.RemoveCommand(c)
	}
	cmdfactory.WorkloadsCmdInit(cmd, api.Runai, api.Training)
}
