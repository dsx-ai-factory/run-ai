/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workspace_test

import (
	"fmt"
	"os"
	"testing"

	policyconstants "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/constants"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/utils"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workspace"

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

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Workspace Tests")
}

const (
	TestFilePath = "workspace"

	ExpectedWorkspaceListWithProjectThatDoesNotExist = "project name: p300, not found\n"
	ExpectSuccessMsg                                 = "Creating workspace %s..."
	MissingProject                                   = "Error: project is missing. value is empty.\n"
	deletingMsg                                      = "Deleting workload '%s'...\n"
	resumingMsg                                      = "Resuming workload '%s'...\nTo track the workload's status, run 'runai workspace describe %s'\n"
	stoppingMsg                                      = "Stopping workload '%s'...\nTo track the workload's status, run 'runai workspace describe %s'\n"
	runAsUserError                                   = "Error: Invalid flags. if [run-as-user] flag is set non of the others [run-as-uid, run-as-gid, user-group-source, supplemental-groups] should be set\n"
)

const (
	WorkspaceMinimalName      = "ws-minimal"
	WorkspaceMinimalImageName = "minimal-image"

	WorkspaceComputeFullName          = "ws-compute-full"
	WorkspaceComputeFullNameImageName = "compute-full-image"
	WorkspacePrefixName               = "workspace-prefix"
)

var workspaceList = []string{WorkspaceMinimalName, WorkspaceComputeFullName, WorkspacePrefixName}

var _ = Describe("Workspace Test", Ordered, func() {
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

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Workspace{}, workspaceList)
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

	Context("Workspace list validation", Ordered, func() {
		It("workspace list all", func() {
			output, err := cli.RunCommand("workspace list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
		})

		It("workspace list all and project flags", func() {
			_, err := cli.RunCommand("workspace list -A --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("[all project] are set none of the others can be;"))
		})

		It("workspace list project p301", func() {
			output, err := cli.RunCommand("workspace list --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))
		})

		It("workspace list project p300", func() {
			output, err := cli.RunCommand("workspace list --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedWorkspaceListWithProjectThatDoesNotExist))
		})

		Describe("Workspace list project flags", func() {
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
				output, err := cli.RunCommand("workspace list", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))

			})

			It("Training list all projects with default project set", func() {
				output, err := cli.RunCommand("workspace list --all", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).To(ContainSubstring(clients.WorkloadName02))
			})
		})

		Describe("Workspace list famework", func() {
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

			It("Workspace with framework - no such flag", func() {
				_, err := cli.RunCommand("workspace list --framework runai", envMap)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unknown flag: --framework"))
			})
		})
	})

	Context("Workspace delete validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})
		It("workspace delete no uuid or name", func() {
			output, err := cli.RunCommand("workspace delete --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to delete workload. one of the attributes must be set name|uuid\n"))
		})

		It("workspace delete both uuid and name", func() {
			output, err := cli.RunCommand("workspace delete anyname --project=p301 --uuid=anothername", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to delete workload. one of the attributes must be set name|uuid\n"))
		})

		It("workspace delete, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace delete --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(deletingMsg, clients.UuidExist)))
		})

		It("workspace delete, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace delete --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to delete workload. generic error string\n"))
		})

		It("workspace delete success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace delete %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(deletingMsg, clients.WorkloadName01)))
		})

		It("workspace delete alias, failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace del %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(deletingMsg, clients.WorkloadName01)))
		})
	})

	Context("Workspace suspend validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("workspace suspend no uuid or name", func() {
			output, err := cli.RunCommand("workspace suspend --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to suspend workload. one of the attributes must be set name|uuid\n"))
		})

		It("workspace suspend both uuid and name", func() {
			output, err := cli.RunCommand("workspace suspend anyname --project=p301 --uuid=anothername", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to suspend workload. one of the attributes must be set name|uuid\n"))
		})

		It("workspace suspend, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace suspend --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(stoppingMsg, clients.UuidExist, clients.UuidExist)))
		})

		It("workspace suspend, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace suspend --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to suspend workload. generic error string\n"))
		})

		It("workspace suspend, name - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace suspend %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(stoppingMsg, clients.WorkloadName01, clients.WorkloadName01)))
		})

		It("workspace suspend, name - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace suspend %s --project=p301", clients.WorkloadName02), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix("Error: failed to suspend workload. generic error string"))
		})
	})

	Context("Workspace resume validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("workspace resume no uuid or name", func() {
			output, err := cli.RunCommand("workspace resume --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to resume workload. one of the attributes must be set name|uuid\n"))
		})

		It("workspace resume both uuid and name", func() {
			output, err := cli.RunCommand("workspace resume anyname --project=p301 --uuid=anothername", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(consts.ErrorNameOrUuidIsNeeded))
			Expect(output).To(Equal("Error: failed to resume workload. one of the attributes must be set name|uuid\n"))
		})

		It("workspace resume, uuid - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace resume --project=p301 --uuid=%s", clients.UuidExist), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(resumingMsg, clients.UuidExist, clients.UuidExist)))
		})

		It("workspace resume, uuid - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace resume --project=p301 --uuid=%s", clients.UuidDontExist), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal("Error: failed to resume workload. generic error string\n"))
		})

		It("workspace resume, name - success", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace resume %s --project=p301", clients.WorkloadName01), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(resumingMsg, clients.WorkloadName01, clients.WorkloadName01)))
		})

		It("workspace resume, name - failed", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace resume %s --project=p301", clients.WorkloadName02), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix("Error: failed to resume workload. generic error string\n"))
		})
	})

	Context("Workspace submit prerequisite validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("workspace submit - no project", func() {
			output, err := cli.RunCommand("workspace submit", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix(MissingProject))
		})

		It("workspace submit - no project", func() {
			output, err := cli.RunCommand("workspace submit anyname", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix(MissingProject))
		})

		It("workspace submit - missing flags", func() {
			output, err := cli.RunCommand("workspace submit --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedWorkspaceListWithProjectThatDoesNotExist))
		})

		It("workspace submit - project does not exist", func() {
			output, err := cli.RunCommand("workspace submit ws01 --project=p300 --image=lies", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedWorkspaceListWithProjectThatDoesNotExist))
		})
	})

	Context("Workspace submit validation", Ordered, func() {
		BeforeEach(func() {
			ResetRunaiCommand()
		})

		It("success naive", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s",
				WorkspaceMinimalName, WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceMinimalName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceMinimalName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))

			Expect(submitSpec.GetCreateHomeDir()).To(Equal(false))
		})

		It("success container", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--command "+
				"--image-pull-policy=IfNotPresent "+
				"--working-dir=/test/path "+
				"--create-home-dir "+
				"--environment-variable=env1=value1 "+
				"--required-pod-topology-key=IS_REQ "+
				"-- this is what I need",
				WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceComputeFullNameImageName))

			Expect(submitSpec.GetCommand()).To(Equal("this is what I need"))
			Expect(submitSpec.GetImagePullPolicy()).To(Equal(openapi.IMAGEPULLPOLICY_IF_NOT_PRESENT))
			Expect(submitSpec.GetWorkingDir()).To(Equal("/test/path"))
			Expect(submitSpec.GetTty()).To(Equal(false))
			Expect(submitSpec.GetStdin()).To(Equal(false))
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
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--cpu-memory-request=2G --cpu-memory-limit=1G "+
				"--cpu-core-limit=0.2 --cpu-core-request=0.3 "+
				"--gpu-devices-request=2 "+
				"--gpu-memory-limit=20M --gpu-memory-request=30M "+
				"--gpu-portion-limit=0.23 --gpu-portion-request=0.43 "+
				"--gpu-request-type=memory "+
				"--large-shm=true",
				WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
			Expect(output).To(ContainSubstring("Flag --gpu-portion-request has been deprecated, use --gpu-memory-portion-request for GPU memory fractions, or --gpu-devices-request for whole-GPU counts"))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceComputeFullNameImageName))
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
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--annotation=ankey1=anvalue1 --annotation=ankey2=anvalue2 "+
				"--label=lkey1=lvalue1 --label=lkey2=lvalue2",
				WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceComputeFullNameImageName))
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
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--configmap-map-volume=name=CONFIGMAP_NAME,path=PATH "+
				"--existing-pvc=claimname=CLAIM_NAME_EXIST,path=PATH_EXIST_PVC "+
				"--new-pvc=claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral "+
				"--git-sync=name=NAME,repository=REPO,path=PATH,secret=SECRET,rev=REVISION "+
				"--host-path=path=PATH,mount=MOUNT,mount-propagation=HostToContainer,readwrite "+
				"--nfs=path=PATH,server=SERVER,mountpath=MOUNT_PATH,readwrite "+
				"--s3=name=NAME,bucket=BUCKET,path=PATH,accesskey=ACCESS_KEY,url=URL",
				WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceComputeFullNameImageName))
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

		It("success network", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--port=service-type=NodePort,container=80,external=8080 "+
				"--external-url=container=9443,url=https://external.runai.com,authusers=user1:user2,authgroups=group1:group3",
				WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceComputeFullNameImageName))
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

		It("success lifecycle", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--auto-deletion-time-after-completion=5h "+
				"--termination-grace-period=3m "+
				"--backoff-limit=9 "+
				"--preemptible",
				WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix("Flag --preemptible has been deprecated, use --preemptibility instead"))
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceComputeFullNameImageName))

			Expect(*submitSpec.AutoDeletionTimeAfterCompletionSeconds.Get()).To(Equal(int64(18000)))
			Expect(*submitSpec.TerminationGracePeriodSeconds.Get()).To(Equal(int64(180)))
			Expect(*submitSpec.BackoffLimit.Get()).To(Equal(int64(9)))
			Expect(*submitSpec.PriorityClass.Get()).To(Equal(policyconstants.PriorityClassInteractivePreemptible))
		})

		Context("success security", func() {
			It("general security flags", func() {
				output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
					"--allow-privilege-escalation "+
					"--run-as-uid=100 "+
					"--run-as-gid=200 "+
					"--user-group-source=fromIdpToken "+
					"--seccomp-profile=Localhost "+
					"--supplemental-groups=123 "+
					"--host-ipc "+
					"--host-network",
					WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
				submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
				Expect(err).To(BeNil())
				Expect(submitSpec.Image.IsSet()).To(Equal(true))
				Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceComputeFullNameImageName))

				security := submitSpec.GetSecurity()
				Expect(security.GetAllowPrivilegeEscalation()).To(Equal(true))
				Expect(security.GetRunAsUid()).To(Equal(int64(100)))
				Expect(security.GetRunAsGid()).To(Equal(int64(200)))
				Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_FROM_IDP_TOKEN))
				Expect(security.GetSeccompProfileType()).To(Equal(openapi.SECCOMPPROFILETYPE_LOCALHOST))
				Expect(security.GetSupplementalGroups()).To(Equal("123"))
				Expect(security.GetHostIpc()).To(Equal(true))
				Expect(security.GetHostNetwork()).To(Equal(true))
			})

			It("security container user flags", func() {
				output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
					"--run-as-uid=300 "+
					"--run-as-gid=400 "+
					"--seccomp-profile=Localhost "+
					"--supplemental-groups=456 --supplemental-groups=123",
					WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
				submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
				Expect(err).To(BeNil())

				security := submitSpec.GetSecurity()
				Expect(security.GetAllowPrivilegeEscalation()).To(Equal(false))
				Expect(security.GetRunAsUid()).To(Equal(int64(300)))
				Expect(security.GetRunAsGid()).To(Equal(int64(400)))
				Expect(security.GetUidGidSource()).To(Equal(openapi.UIDGIDSOURCE_CUSTOM))
				Expect(security.GetSeccompProfileType()).To(Equal(openapi.SECCOMPPROFILETYPE_LOCALHOST))
				Expect(security.GetSupplementalGroups()).To(Equal("456,123"))
				Expect(security.GetHostIpc()).To(Equal(false))
				Expect(security.GetHostNetwork()).To(Equal(false))
			})

			Context("run-as-user security flags", func() {
				BeforeEach(func() {
					ResetRunaiCommand()
				})

				It("should pass", func() {
					output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
						"--run-as-user",
						WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
					Expect(err).To(BeNil())
					Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))

					submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
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
					output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
						"--run-as-uid=300 "+
						"--run-as-user",
						WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
				It("should failed when using mutual exclusive flags - run-as-gid", func() {
					output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
						"--run-as-gid=200 "+
						"--run-as-user",
						WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
				It("should failed when using mutual exclusive flags - supplemental-groups", func() {
					output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
						"--supplemental-groups=1,2,3 "+
						"--run-as-user",
						WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
				It("should failed when using mutual exclusive flags - user-group-source", func() {
					output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
						"--user-group-source=fromIdpToken "+
						"--run-as-user",
						WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
					Expect(err).ToNot(BeNil())
					Expect(output).To(Equal(runAsUserError))
				})
			})

			It("capabilities", func() {
				output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
					"--capability SYS_TTY_CONFIG "+
					"--capability LINUX_IMMUTABLE "+
					"--capability NET_BIND_SERVICE "+
					"--capability IPC_OWNER",
					WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
				submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
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
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--node-pools=node-one --node-pools=node-two "+
				"--toleration=name=NAME,operator=Exists,key=KEY,effect=PreferNoSchedule "+
				"--node-type=the-best-one",
				WorkspaceComputeFullName, WorkspaceComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspaceComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceComputeFullName)
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

		It("prefix prefix - generate workspace with prefix name", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit --project=p301 --image=%s --name-prefix=%s",
				WorkspaceMinimalImageName, WorkspacePrefixName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspacePrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspacePrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))
		})

		It("execute string at the end goes to args", func() {
			output, err := cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s -- this is args",
				WorkspacePrefixName, WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(fmt.Sprintf(ExpectSuccessMsg, WorkspacePrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspacePrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))

			Expect(submitSpec.GetArgs()).To(Equal("this is args"))
		})

	})
})

var _ = Describe("Workspace Invalid Test", Ordered, func() {
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
		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Workspace{}, workspaceList)
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

	Context("Workspace submit", Ordered, func() {
		It("completions doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--completions",
				WorkspaceMinimalName, WorkspaceMinimalImageName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --completions"))
		})

		It("parallelism doesn't exist ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s "+
				"--parallelism",
				WorkspaceMinimalName, WorkspaceMinimalImageName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown flag: --parallelism"))
		})

		It("cannot have args and command together ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s --name-prefix=shouldnotwork",
				WorkspaceMinimalName, WorkspaceMinimalImageName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("both prefix and name are set, only one can be used"))
		})

		It("cannot have --priority --preemptible together ", func() {
			_, err = cli.RunCommand(fmt.Sprintf("workspace submit %s --project=p301 --image=%s --priority=build --preemptible",
				WorkspaceMinimalName, WorkspaceMinimalImageName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("if any flags in the group [priority preemptible] are set none of the others can be; [preemptible priority] were all set"))
		})
	})
})

func ResetRunaiCommand() {
	cmd := workspace.Cmd
	for _, c := range cmd.Commands() {
		cmd.RemoveCommand(c)
	}
	cmdfactory.WorkloadsCmdInit(cmd, api.Runai, api.Workspace)
}
