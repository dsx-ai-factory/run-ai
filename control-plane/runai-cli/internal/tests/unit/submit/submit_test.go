/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	submitcmd "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/submit"
	submitapi "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/submit"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/submit"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Submit Tests")
}

const (
	TestFilePath = "submit"

	ExpectedWorkspaceListWithProjectThatDoesNotExist = "project name: p300, not found\n"
	ExpectWorkspaceSuccessMsg                        = "Creating workspace %s..."
	ExpectWorkspaceDepMsg                            = "Command \"submit\" is deprecated, please use: `runai workspace submit` or `runai training submit` instead"
	ExpectTrainingDepMsg                             = "Command \"submit\" is deprecated, please use: `runai workspace submit` or `runai training submit` instead\n"
	ExpectTrainingSuccessMsg                         = "Creating training %s..."
	MissingProject                                   = "Error: project is missing. value is empty.\n"
)

const (
	WorkspaceMinimalName             = "ws-minimal"
	WorkspaceMinimalImageName        = "minimal-image"
	NoNameWorkspace                  = "workspace"
	NoNameTraining                   = "runai"
	WorkspaceComputeFullName         = "ws-compute-full"
	WorkspacePrefixName              = "workspace-prefix"
	TrainingPrefixName               = "training-prefix"
	TrainingComputeFullName          = "tr-comp-full"
	TrainingComputeFullNameImageName = "tr-comp-image"
)

var workspaceList = []string{NoNameWorkspace, WorkspacePrefixName}
var trainingList = []string{WorkspaceMinimalName, WorkspaceComputeFullName, NoNameTraining, TrainingPrefixName, TrainingComputeFullName}

var _ = Describe("Submit Test (training)", Ordered, func() {
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

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Training{}, trainingList)
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

	Context("Submit command flags compare", Ordered, func() {
		It("compare submit --> workspace submit to ", func() {
			outputSubmit, err := cli.RunCommand("submit -h", envMap)
			Expect(err).To(BeNil())
			outputWorkspace, err := cli.RunCommand("workspace submit -h", envMap)
			Expect(err).To(BeNil())
			dcS2W := submit.NewDiffChecker(outputSubmit, outputWorkspace, []string{"--completions int32", "--parallelism int32", "--interactive"})
			Expect(dcS2W.Validate()).To(BeTrue())
		})

		It("compare submit --> training submit to ", func() {
			outputSubmit, err := cli.RunCommand("submit -h", envMap)
			Expect(err).To(BeNil())
			outputTraining, err := cli.RunCommand("training submit -h", envMap)
			Expect(err).To(BeNil())
			dcS2T := submit.NewDiffChecker(outputSubmit, outputTraining, []string{"--preemptible", "--interactive"})
			Expect(dcS2T.Validate()).To(BeTrue())
		})
	})

	Context("Submit prerequisite validation", Ordered, func() {
		It("submit - no project", func() {
			output, err := cli.RunCommand("submit", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(MissingProject))
		})

		It("Submit - missing flags", func() {
			output, err := cli.RunCommand("submit --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("required flag(s) \"image\" not set"))
			Expect(output).To(Equal(ExpectTrainingDepMsg))
		})

		It("Submit - missing flags", func() {
			output, err := cli.RunCommand("submit --project=p300 -i any", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedWorkspaceListWithProjectThatDoesNotExist))
		})

		It("Submit - project does not exist", func() {
			output, err := cli.RunCommand("submit ws01 --project=p300 --image=lies", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedWorkspaceListWithProjectThatDoesNotExist))
		})
	})

	Context("Submit validation", Ordered, func() {
		It("Submit - success naive", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit %s --project=p301 --image=%s",
				WorkspaceMinimalName, WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectTrainingSuccessMsg, WorkspaceMinimalName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspaceMinimalName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))

			Expect(submitSpec.GetCreateHomeDir()).To(Equal(false))
		})

		It("submit - success compute (validate gpu flag)", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit %s --project=p301 --image=%s "+
				"--memory=2G --memory-limit=1G "+
				"--cpu-limit=0.2 --cpu=0.3 "+
				"--gpu=2 "+
				"--gpu-memory=30M "+
				"--large-shm=true",
				TrainingComputeFullName, TrainingComputeFullNameImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectTrainingSuccessMsg, TrainingComputeFullName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingComputeFullName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(TrainingComputeFullNameImageName))
			Expect(*submitSpec.Compute.Get()).ToNot(BeNil())

			compute := submitSpec.Compute.Get()
			Expect(*compute.CpuMemoryRequest.Get()).To(Equal("2G"))
			Expect(*compute.CpuMemoryLimit.Get()).To(Equal("1G"))
			Expect(compute.GetGpuDevicesRequest()).To(Equal(int32(2)))
			Expect(*compute.GpuMemoryRequest.Get()).To(Equal("30M"))
			Expect(compute.HasGpuMemoryLimit()).To(Equal(false))
			Expect(compute.HasGpuPortionRequest()).To(Equal(false))
			Expect(compute.HasGpuPortionLimit()).To(Equal(false))
			Expect(compute.HasGpuRequestType()).To(Equal(false))
			Expect(*compute.LargeShmRequest.Get()).To(Equal(true))
		})
	})

	Context("Submit validation - backward compatibility", Ordered, func() {
		It("Submit - no name - generate workspace with prefix enabled", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit --project=p301 --image=%s",
				WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectTrainingSuccessMsg, NoNameTraining)))
			submitSpec, err := clients.GetSpecByWorkspaceName(NoNameTraining)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))

			Expect(submitSpec.GetCreateHomeDir()).To(Equal(false))
		})

		It("Submit - execute string at the end goes to args", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit --project=p301 --image=%s -- this is args",
				WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectTrainingSuccessMsg, NoNameTraining)))
			submitSpec, err := clients.GetSpecByWorkspaceName(NoNameTraining)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))

			Expect(submitSpec.GetArgs()).To(Equal("this is args"))
		})

		It("Submit with command flag - execute string at the end goes to command", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit --project=p301 --command --image=%s -- this is command",
				WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectTrainingSuccessMsg, NoNameTraining)))
			submitSpec, err := clients.GetSpecByWorkspaceName(NoNameTraining)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))
			Expect(submitSpec.GetCommand()).To(Equal("this is command"))
		})

		It("Submit - job name prefix - generate training with prefix name", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit --project=p301 --image=%s --job-name-prefix=%s",
				WorkspaceMinimalImageName, TrainingPrefixName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectTrainingSuccessMsg, TrainingPrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(TrainingPrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))
		})
	})
})

var _ = Describe("Submit Test (workspace)", Ordered, func() {
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
	})

	Context("Submit validation - backward compatibility", Ordered, func() {
		It("Submit - no name - generate workspace with prefix enabled", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit --interactive --project=p301 --image=%s",
				WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(ExpectWorkspaceDepMsg))
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectWorkspaceSuccessMsg, NoNameWorkspace)))
			submitSpec, err := clients.GetSpecByWorkspaceName(NoNameWorkspace)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))

			Expect(submitSpec.GetCreateHomeDir()).To(Equal(false))
		})

		It("Submit - job name prefix - generate workspace with prefix name", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit --interactive --project=p301 --image=%s --job-name-prefix=%s",
				WorkspaceMinimalImageName, WorkspacePrefixName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectWorkspaceSuccessMsg, WorkspacePrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspacePrefixName)
			Expect(err).To(BeNil())
			Expect(submitSpec.Image.IsSet()).To(Equal(true))
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))
		})
	})
})

var _ = Describe("Submit Test (backward compatibility)", Ordered, func() {
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
		submitcmd.SubmitSpec.Clear()
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
	})

	Context("Submit validation - backward compatibility validation", Ordered, func() {
		It("Submit gpu flag validation", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit --interactive --project=p301 --image=%s -g=0.4",
				WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectWorkspaceSuccessMsg, NoNameWorkspace)))
			submitSpec, err := clients.GetSpecByWorkspaceName(NoNameWorkspace)
			Expect(err).To(BeNil())

			c := submitSpec.GetCompute()
			Expect(c.GetGpuPortionRequest()).To(Equal(0.4))
			Expect(c.HasGpuDevicesRequest()).To(Equal(false))

			output, err = cli.RunCommand(fmt.Sprintf("submit --interactive --project=p301 --image=%s -g=2",
				WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectWorkspaceSuccessMsg, NoNameWorkspace)))
			submitSpec, err = clients.GetSpecByWorkspaceName(NoNameWorkspace)
			Expect(err).To(BeNil())

			c = submitSpec.GetCompute()
			Expect(c.HasGpuPortionRequest()).To(Equal(false))
			Expect(c.GetGpuDevicesRequest()).To(Equal(int32(2)))

			output, err = cli.RunCommand(fmt.Sprintf("submit --interactive --project=p301 --image=%s -g=2.3",
				WorkspaceMinimalImageName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix("Error: failed to submit. gpu value cannot be a fraction in whole numbers: 2.30\n"))
		})

		It("Submit - command contain groups - command and annotation contain groups", func() {
			args := strings.Split(fmt.Sprintf("submit --job-name-prefix %s -i %s -p p301 --interactive --command -- /bin/bash -c",
				WorkspacePrefixName, WorkspaceMinimalImageName), " ")
			args = append(args, "cat /etc/os-release;echo $PATH && pwd")
			output, err := cli.RunCommandStringSplit(context.TODO(), args, envMap)

			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectWorkspaceSuccessMsg, WorkspacePrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspacePrefixName)
			Expect(err).To(BeNil())

			Expect(submitSpec.GetCommand()).To(Equal("/bin/bash -c 'cat /etc/os-release;echo $PATH && pwd'"))
			Expect(submitSpec.GetAnnotations()[0].GetValue()).To(HaveSuffix("/bin/bash -c 'cat /etc/os-release;echo $PATH && pwd'"))
		})

		It("Submit - no name - generate workspace with prefix enabled", func() {
			output, err := cli.RunCommand(fmt.Sprintf("submit -p p301 "+
				"--job-name-prefix %s "+
				"--interactive "+
				"--add-capability SYS_ADMIN "+
				"--annotation my-ann-key=my-ann-value "+
				"--command "+
				"--create-home-dir "+
				"-e my-env-key=my-env-value "+
				"-i %s "+
				"--image-pull-policy Always "+
				"--label my-label-key=my-label-val "+
				"--preferred-pod-topology-key my-pod-topo-key "+
				// "--stdin --tty "+ // TODO not yet supported
				"--working-dir my-working-dir "+
				"--cpu 0.2 "+
				"--cpu-limit 0.4 "+
				"--extended-resource my-ext-res=4 "+
				"-g 0.3 "+
				"--memory 500M "+
				"--memory-limit 1G "+
				"--configmap-volume name=my-cm,path=/home/mycm "+
				"--existing-pvc claimname=my-claim,path=/home/existing-pvc "+
				"--git-sync source=http://google.com,branch=master,target=/home/target "+
				"--large-shm "+
				"--new-pvc claimname=new-pvc,storageclass=my-sc,size=1G,path=/home/new-pvc,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral "+
				"--s3 bucket=genome-idx,key=,secret=,url=https://s3.us-east-1.amazonaws.com/,target=/s3 "+
				"--s3 bucket=genome-idx-secret,key=s3key,secret=s3secret,url=https://s3.us-east-1.amazonaws.com/,target=/s3cert "+
				"-v /hostpath:/hostpath:ro --volume /nfs:/nfs::1.2.3.4 "+
				"--host-ipc --host-network --backoff-limit 10 --preemptible --allow-privilege-escalation --node-pools default "+
				"--node-type my-node-type "+
				"--toleration operator=Equal,key=my-key,value=my-value,effect=NoSchedule "+
				"-- submit command string",
				WorkspacePrefixName, WorkspaceMinimalImageName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf(ExpectWorkspaceSuccessMsg, WorkspacePrefixName)))
			submitSpec, err := clients.GetSpecByWorkspaceName(WorkspacePrefixName)
			Expect(err).To(BeNil())

			// Spec validation
			// --add-capability SYS_ADMIN
			security := submitSpec.GetSecurity()
			capabilities := security.GetCapabilities()
			Expect(len(capabilities)).To(Equal(1))
			Expect(capabilities[0]).To(Equal(openapi.CAPABILITY_SYS_ADMIN))

			// -e my-env-key=my-env-value
			ev := submitSpec.GetEnvironmentVariables()
			Expect(len(ev)).To(Equal(1))
			Expect(ev[0].GetName()).To(Equal("my-env-key"))
			Expect(ev[0].GetValue()).To(Equal("my-env-value"))

			// --annotation my-ann-key=my-ann-value + generated runai-submit-command
			annotations := submitSpec.GetAnnotations()
			Expect(len(annotations)).To(Equal(2))
			Expect(annotations[0].GetName()).To(Equal("my-ann-key"))
			Expect(annotations[0].GetValue()).To(Equal("my-ann-value"))
			Expect(annotations[1].GetName()).To(Equal("runai-submit-command"))

			// --label my-label-key=my-label-val
			labels := submitSpec.GetLabels()
			Expect(len(labels)).To(Equal(1))
			Expect(labels[0].GetName()).To(Equal("my-label-key"))
			Expect(labels[0].GetValue()).To(Equal("my-label-val"))

			// --preferred-pod-topology-key my-pod-topo-key
			pa := submitSpec.GetPodAffinity()
			Expect(pa.GetType()).To(Equal(openapi.PODAFFINITYTYPE_PREFERRED))
			Expect(pa.GetKey()).To(Equal("my-pod-topo-key"))

			Expect(submitSpec.GetCommand()).To(Equal("submit command string"))

			// --image-pull-policy Always + --job-name-prefix = "minimal-image"
			Expect(*submitSpec.Image.Get()).To(Equal(WorkspaceMinimalImageName))
			Expect(*submitSpec.ImagePullPolicy.Get()).To(Equal(openapi.IMAGEPULLPOLICY_ALWAYS))

			// --create-home-dir + --working-dir my-working-dir
			Expect(submitSpec.GetCreateHomeDir()).To(Equal(true))
			Expect(submitSpec.GetWorkingDir()).To(Equal("my-working-dir"))

			// --cpu 0.2 --cpu-limit 0.4
			c := submitSpec.GetCompute()
			Expect(c.GetCpuCoreRequest()).To(Equal(0.2))

			Expect(c.GetCpuCoreLimit()).To(Equal(0.4))
			// -g 0.3 --memory 500M --memory-limit 1G
			Expect(c.GetGpuPortionRequest()).To(Equal(0.3))
			Expect(c.GetCpuMemoryRequest()).To(Equal("500M"))
			Expect(c.GetCpuMemoryLimit()).To(Equal("1G"))

			// --large-shm
			Expect(c.GetLargeShmRequest()).To(Equal(true))

			// --extended-resource my-ext-res=4
			er := c.GetExtendedResources()
			Expect(len(er)).To(Equal(1))
			Expect(er[0].GetResource()).To(Equal("my-ext-res"))
			Expect(er[0].GetQuantity()).To(Equal("4"))

			// --configmap-volume name=my-cm,path=/home/mycm
			s := submitSpec.GetStorage()
			cm := s.GetConfigMapVolume()
			Expect(len(cm)).To(Equal(1))
			Expect(cm[0].GetConfigMap()).To(Equal("my-cm"))
			Expect(cm[0].GetMountPath()).To(Equal("/home/mycm"))

			// --new-pvc claimname=new-pvc,storageclass=my-sc,size=1G,path=/home/new-pvc,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral
			// --existing-pvc claimname=my-claim,path=/home/existing-pvc
			pvc := s.GetPvc()
			Expect(len(pvc)).To(Equal(2))
			Expect(pvc[0].GetExistingPvc()).To(Equal(false))
			Expect(pvc[0].GetClaimName()).To(Equal("new-pvc"))
			Expect(pvc[0].GetPath()).To(Equal("/home/new-pvc"))
			Expect(pvc[0].GetReadOnly()).To(Equal(true))
			Expect(pvc[0].GetEphemeral()).To(Equal(true))
			ci := pvc[0].GetClaimInfo()
			Expect(ci.GetSize()).To(Equal("1G"))
			Expect(ci.GetStorageClass()).To(Equal("my-sc"))
			am := ci.GetAccessModes()
			Expect(am.GetReadWriteMany()).To(Equal(true))
			Expect(am.GetReadOnlyMany()).To(Equal(true))
			Expect(am.GetReadWriteOnce()).To(Equal(true))

			Expect(pvc[1].GetExistingPvc()).To(Equal(true))
			Expect(pvc[1].GetClaimName()).To(Equal("my-claim"))
			Expect(pvc[1].GetPath()).To(Equal("/home/existing-pvc"))

			// --git-sync source=http://google.com,branch=master,target=/home/target
			git := s.GetGit()
			Expect(len(git)).To(Equal(1))
			Expect(git[0].GetRepository()).To(Equal("http://google.com"))
			Expect(git[0].GetRevision()).To(Equal("master"))
			Expect(git[0].GetPath()).To(Equal("/home/target"))

			// --s3 bucket=genome-idx,key=,secret=,url=https://s3.us-east-1.amazonaws.com/,target=/s3
			// --s3 bucket=genome-idx-secret,key=s3key,secret=s3secret,url=https://s3.us-east-1.amazonaws.com/,target=/s3cert
			s3 := s.GetS3()
			Expect(len(s3)).To(Equal(2))
			Expect(s3[0].GetBucket()).To(Equal("genome-idx"))
			Expect(s3[0].GetPath()).To(Equal("/s3"))
			Expect(s3[0].GetUrl()).To(Equal("https://s3.us-east-1.amazonaws.com/"))

			Expect(s3[1].GetBucket()).To(Equal("genome-idx-secret"))
			Expect(s3[1].GetPath()).To(Equal("/s3cert"))
			Expect(s3[1].GetUrl()).To(Equal("https://s3.us-east-1.amazonaws.com/"))
			Expect(s3[1].GetSecretKeyOfAccessKeyId()).To(Equal("s3key"))
			Expect(s3[1].GetSecretKeyOfSecretKey()).To(Equal("s3secret"))

			// -v /hostpath:/hostpath:ro --volume /nfs:/nfs::1.2.3.4
			nfs := s.GetNfs()
			Expect(len(nfs)).To(Equal(1))
			Expect(nfs[0].GetPath()).To(Equal("/nfs"))
			Expect(nfs[0].GetMountPath()).To(Equal("/nfs"))
			Expect(nfs[0].GetServer()).To(Equal("1.2.3.4"))
			hp := s.GetHostPath()
			Expect(len(hp)).To(Equal(1))
			Expect(hp[0].GetPath()).To(Equal("/hostpath"))
			Expect(hp[0].GetMountPath()).To(Equal("/hostpath"))

			// --toleration operator=Equal,key=my-key,value=my-value,effect=NoSchedule
			t := submitSpec.Tolerations
			Expect(len(t)).To(Equal(1))
			Expect(t[0].GetName()).To(HavePrefix("workspace-"))
			Expect(t[0].GetOperator()).To(Equal(openapi.TOLERATIONOPERATOR_EQUAL))
			Expect(t[0].GetValue()).To(Equal("my-value"))
			Expect(t[0].GetEffect()).To(Equal(openapi.TOLERATIONEFFECT_NO_SCHEDULE))
		})
	})
})

var _ = Describe("Serving Port Parsing", func() {
	Context("Standard Inference", func() {
		It("parse plain port number", func() {
			result, err := submitapi.ParseServingPort("8080")
			Expect(err).To(BeNil())
			Expect(*result.Container.Get()).To(Equal(int32(8080)))
		})

		It("parse full format with cluster-local-access-only", func() {
			result, err := submitapi.ParseServingPort("container=9090,protocol=http,cluster-local-access-only")
			Expect(err).To(BeNil())
			Expect(*result.Container.Get()).To(Equal(int32(9090)))
			Expect(*result.ClusterLocalAccessOnly.Get()).To(BeTrue())
		})

		It("reject invalid port", func() {
			_, err := submitapi.ParseServingPort("99999")
			Expect(err).ToNot(BeNil())
		})

		It("reject unknown key", func() {
			_, err := submitapi.ParseServingPort("container=8080,unknown=value")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("Distributed Inference", func() {
		It("parse plain port number", func() {
			result, err := submitapi.ParseDistributedInferenceServingPort("8080")
			Expect(err).To(BeNil())
			Expect(*result.Port.Get()).To(Equal(int32(8080)))
		})

		It("parse full format with expose-externally", func() {
			result, err := submitapi.ParseDistributedInferenceServingPort("container=9090,protocol=http,expose-externally")
			Expect(err).To(BeNil())
			Expect(*result.Port.Get()).To(Equal(int32(9090)))
			Expect(*result.ExposeExternally.Get()).To(BeTrue())
		})

		It("parse expose-externally=false", func() {
			result, err := submitapi.ParseDistributedInferenceServingPort("container=9090,protocol=http,expose-externally=false")
			Expect(err).To(BeNil())
			Expect(*result.ExposeExternally.Get()).To(BeFalse())
		})

		It("parse expose-externally=true", func() {
			result, err := submitapi.ParseDistributedInferenceServingPort("container=9090,protocol=http,expose-externally=true")
			Expect(err).To(BeNil())
			Expect(*result.ExposeExternally.Get()).To(BeTrue())
		})

		It("default expose-externally to true when omitted", func() {
			result, err := submitapi.ParseDistributedInferenceServingPort("container=9090,protocol=http")
			Expect(err).To(BeNil())
			Expect(*result.ExposeExternally.Get()).To(BeTrue())
		})

		It("reject invalid expose-externally value", func() {
			_, err := submitapi.ParseDistributedInferenceServingPort("container=9090,expose-externally=invalid")
			Expect(err).To(MatchError("invalid value: the key 'expose-externally' inside '--serving-port' must be 'true' or 'false', got 'invalid'"))
		})

		It("parse with exposed-url", func() {
			result, err := submitapi.ParseDistributedInferenceServingPort("container=8080,expose-externally,exposed-url=my-custom-endpoint")
			Expect(err).To(BeNil())
			Expect(*result.Port.Get()).To(Equal(int32(8080)))
			Expect(*result.ExposedUrl.Get()).To(Equal("my-custom-endpoint"))
		})

		It("reject invalid port", func() {
			_, err := submitapi.ParseDistributedInferenceServingPort("99999")
			Expect(err).ToNot(BeNil())
		})

		It("reject unknown key", func() {
			_, err := submitapi.ParseDistributedInferenceServingPort("container=8080,unknown=value")
			Expect(err).ToNot(BeNil())
		})
	})
})
