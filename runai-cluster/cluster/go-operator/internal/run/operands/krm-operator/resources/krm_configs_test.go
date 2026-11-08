/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

func defaultRunaiConfig() *runaiv1.RunaiConfig {
	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai"},
	}
	runaiConfig.SetDefaultsWhereNeeded()
	return runaiConfig
}

// openShiftSCCBindings are the SCC bindings added only on OpenShift.
var openShiftSCCBindings = []string{
	"runai-jupyter-notebook-scc",
	"runai-openshift-s3-scc",
	"runai-user-job",
}

func krmConfigFor(runaiConfig *runaiv1.RunaiConfig) *kaiv1alpha1.KRMConfig {
	Expect(kaiv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	k8sClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

	resource, err := KRMConfigForRunaiConfig(context.TODO(), k8sClient, runaiConfig)
	Expect(err).NotTo(HaveOccurred())
	Expect(resource).NotTo(BeNil())

	return resource.(*kaiv1alpha1.KRMConfig)
}

var _ = Describe("KRMConfigForRunaiConfig", func() {
	It("names the singleton and takes the namespace from the RunaiConfig", func() {
		krmConfig := krmConfigFor(defaultRunaiConfig())

		Expect(krmConfig.GetName()).To(Equal(kaiv1alpha1.KRMConfigSingletonName))
		Expect(krmConfig.Spec.Namespace).To(Equal("runai"))
	})

	It("pins the vocabulary the scheduler was installed with", func() {
		global := krmConfigFor(defaultRunaiConfig()).Spec.Global
		Expect(global).NotTo(BeNil())

		Expect(global.SchedulerName).To(Equal(ptr.To(clustercommonconstants.RunaiScheduler)))
		Expect(global.QueueLabelKey).To(Equal(ptr.To(clustercommonconstants.RunaiQueueLabel)))
		Expect(global.NodePoolLabelKey).To(Equal(ptr.To(clustercommonconstants.NodePoolNameLabel)))
		// Deliberately the same key as QueueLabelKey.
		Expect(global.NamespaceProjectLabelKey).To(Equal(ptr.To(clustercommonconstants.RunaiQueueLabel)))
		Expect(global.ProjectLabelKey).To(Equal(ptr.To(clustercommonconstants.RunaiProjectLabel)))
		Expect(global.FinalizerDomain).To(Equal(ptr.To(clustercommonconstants.RunaiOrgDomain)))
		Expect(global.EnforceSchedulerAnnotationKey).
			To(Equal(ptr.To(clustercommonconstants.EnforceSchedulerAnnotationName)))
	})

	// The mode reaches KRM only through this CR: the chart renders no KRMConfig
	// in external mode, so a mode that does not survive the conversion leaves
	// every KRM-deployed service on the image default instead of the requested one.
	DescribeTable("carries the FIPS mode onto the CR",
		func(fipsMode string, expected *kaiv1alpha1.FipsMode) {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.Internal.Global.FipsMode = fipsMode

			Expect(krmConfigFor(runaiConfig).Spec.Global.FipsMode).To(Equal(expected))
		},
		Entry("off", "off", ptr.To(kaiv1alpha1.FipsModeOff)),
		Entry("on", "on", ptr.To(kaiv1alpha1.FipsModeOn)),
		Entry("only", "only", ptr.To(kaiv1alpha1.FipsModeOnly)),
		// Left unset so KRM applies its own default, rather than pinning a mode
		// the services cannot honour.
		Entry("unset", "", nil),
		Entry("unrecognised", "enabled", nil),
	)

	It("leaves leader election to KRM", func() {
		Expect(krmConfigFor(defaultRunaiConfig()).Spec.Global.LeaderElection).To(BeNil())
	})

	It("asks for both service monitors", func() {
		serviceMonitor := krmConfigFor(defaultRunaiConfig()).Spec.Global.ServiceMonitor
		Expect(serviceMonitor).NotTo(BeNil())
		Expect(serviceMonitor.Enabled).To(Equal(ptr.To(true)))
		Expect(serviceMonitor.Accounting).To(Equal(ptr.To(true)))
	})

	Context("project-controller", func() {
		It("turns both validating webhooks off, since the agent serves them", func() {
			projectController := krmConfigFor(defaultRunaiConfig()).Spec.ProjectController
			Expect(projectController).NotTo(BeNil())
			Expect(projectController.Webhooks.EnableProjectValidation).To(Equal(ptr.To(false)))
			Expect(projectController.Webhooks.EnableDepartmentValidation).To(Equal(ptr.To(false)))
			Expect(projectController.ControllerService.Webhook).To(BeNil())
		})

		It("carries the per-project feature toggles", func() {
			features := krmConfigFor(defaultRunaiConfig()).Spec.ProjectController.Features
			Expect(features.CreateNamespaces).To(Equal(ptr.To(true)))
			Expect(features.CreateRoleBindings).To(Equal(ptr.To(true)))
			Expect(features.LimitRange).To(Equal(ptr.To(false)))
		})

		It("carries the runai label vocabulary", func() {
			args := krmConfigFor(defaultRunaiConfig()).Spec.ProjectController.Args
			Expect(args.ProjectNamePrefix).To(Equal(ptr.To(clustercommonconstants.NamespaceRunAi)))
			Expect(args.ProjectIDLabelKey).To(Equal(ptr.To(clustercommonconstants.ProjectIdLabel)))
			Expect(args.QueueDepartmentNameLabelKey).
				To(Equal(ptr.To(clustercommonconstants.QueueDepartmentNameLabel)))
			Expect(args.LimitRangeName).To(Equal(ptr.To(clustercommonconstants.RunaiLimitRangeName)))
		})

		It("translates verbosity into the debug flag", func() {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.ProjectController.Args.Verbosity = ptr.To(-1)
			Expect(krmConfigFor(runaiConfig).Spec.ProjectController.Args.Debug).To(Equal(ptr.To(true)))

			runaiConfig.Spec.ProjectController.Args.Verbosity = ptr.To(0)
			Expect(krmConfigFor(runaiConfig).Spec.ProjectController.Args.Debug).To(Equal(ptr.To(false)))
		})

		It("carries the profiler port as a number", func() {
			profiling := krmConfigFor(defaultRunaiConfig()).Spec.ProjectController.Profiling
			Expect(profiling).NotTo(BeNil())
			Expect(profiling.APIPort).NotTo(BeNil())
		})

		It("carries every deletion blocker onto the CR", func() {
			blockers := krmConfigFor(defaultRunaiConfig()).Spec.ProjectController.DeleteBlockers
			Expect(blockers).To(HaveLen(12))

			var displayNames []string
			for _, blocker := range blockers {
				displayNames = append(displayNames, blocker.DisplayName)
			}
			Expect(displayNames).To(ContainElements("Workloads", "DataVolumes", "Secrets", "Pvcs"))
		})

		It("scopes the Secret and PVC blockers to project-scoped objects", func() {
			blockers := krmConfigFor(defaultRunaiConfig()).Spec.ProjectController.DeleteBlockers
			for _, blocker := range blockers {
				if blocker.Kind == "Secret" || blocker.Kind == "PersistentVolumeClaim" {
					Expect(blocker.LabelSelector).NotTo(BeNil())
				}
			}
		})

		It("carries the per-project role bindings KRM now owns", func() {
			projectController := krmConfigFor(defaultRunaiConfig()).Spec.ProjectController
			Expect(projectController.ExtraProjectRoleBindings).To(HaveLen(18))

			var names []string
			for _, binding := range projectController.ExtraProjectRoleBindings {
				names = append(names, binding.Name)
				Expect(binding.ServiceAccountName).NotTo(BeEmpty())
				// Everything outside the OpenShift set binds the install namespace.
				Expect(binding.BindProjectServiceAccount).To(BeNil())
			}
			Expect(names).To(ContainElements("runai-agent", "cluster-sync", "anyworkload-controller", "inference-endpoint-controller"))
			// KRM's controller reaches project namespaces cluster-wide, and the Run:ai
			// service account this bound no longer exists.
			Expect(names).NotTo(ContainElement("runai-project-controller-project"))
			Expect(names).NotTo(ContainElements(openShiftSCCBindings))
		})

		It("adds the OpenShift SCC bindings against the project's own service account", func() {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runv1internal.KubernetesFlavorOpenShift

			bindings := krmConfigFor(runaiConfig).Spec.ProjectController.ExtraProjectRoleBindings
			Expect(bindings).To(HaveLen(21))

			found := map[string]kaiv1alpha1.ProjectRoleBinding{}
			for _, binding := range bindings {
				found[binding.Name] = binding
			}
			for _, name := range openShiftSCCBindings {
				Expect(found).To(HaveKey(name))
				// Jupyter, S3 and hostpath all need the project namespace's own
				// default account, not the install namespace's.
				Expect(found[name].ServiceAccountName).To(Equal("default"))
				Expect(found[name].BindProjectServiceAccount).To(HaveValue(BeTrue()))
				// KRM defaults ClusterRoleName to Name.
				Expect(found[name].ClusterRoleName).To(BeEmpty())
			}
		})

		It("leaves the limit range binding to KRM's builtin", func() {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.ProjectController.LimitRange = ptr.To(true)

			var names []string
			for _, b := range krmConfigFor(runaiConfig).Spec.ProjectController.ExtraProjectRoleBindings {
				names = append(names, b.Name)
			}
			Expect(names).NotTo(ContainElement(ContainSubstring("limit-range")))
		})

		It("leaves the role bindings ConfigMap name to KRM", func() {
			Expect(krmConfigFor(defaultRunaiConfig()).Spec.ProjectController.RoleBindingsConfigMapName).To(BeNil())
		})

		It("passes no extra args, since KRM's operator derives them from the CR", func() {
			Expect(krmConfigFor(defaultRunaiConfig()).Spec.ProjectController.ExtraArgs).To(BeEmpty())
		})
	})

	Context("nodepool-controller", func() {
		It("turns the NodePool validating webhook off", func() {
			nodePoolController := krmConfigFor(defaultRunaiConfig()).Spec.NodePoolController
			Expect(nodePoolController).NotTo(BeNil())
			Expect(nodePoolController.Webhooks.EnableNodePoolValidation).To(Equal(ptr.To(false)))
			Expect(nodePoolController.ControllerService.Webhook).To(BeNil())
		})

		It("names both published ports", func() {
			controllerService := krmConfigFor(defaultRunaiConfig()).Spec.NodePoolController.ControllerService
			Expect(controllerService.Metrics.Name).To(Equal(ptr.To("metrics")))
			Expect(controllerService.Metrics.Port).To(Equal(ptr.To(int32(9400))))
			Expect(controllerService.NodePoolMetrics.Name).To(Equal(ptr.To("np-metrics")))
			Expect(controllerService.NodePoolMetrics.Port).To(Equal(ptr.To(int32(8080))))
		})

		It("carries the runai vocabulary the binary does not default to", func() {
			args := krmConfigFor(defaultRunaiConfig()).Spec.NodePoolController.Args
			Expect(args.MetricsNamespace).To(Equal(ptr.To(clustercommonconstants.NamespaceRunAi)))
			Expect(args.ExcludedNodepoolName).To(Equal(ptr.To(clustercommonconstants.ExcludedNodepoolName)))
			Expect(args.ManagedNodesConfigName).To(Equal(ptr.To(clustercommonconstants.ManagedNodesConfigName)))
			Expect(args.ToExcludeLabel).To(Equal(ptr.To(clustercommonconstants.ShouldBeExcludedLabelKey)))
			Expect(args.SchedulingShardArgs).NotTo(BeNil())
		})

		It("passes the flags KRMConfig does not model as extra args", func() {
			nodePoolController := krmConfigFor(defaultRunaiConfig()).Spec.NodePoolController
			Expect(nodePoolController.ExtraArgs).To(ContainElements(
				"--scheduler-namespace=runai",
				"--cpu-worker-node-label-key=node-role.kubernetes.io/runai-cpu-worker",
				"--gpu-worker-node-label-key=node-role.kubernetes.io/runai-gpu-worker",
				"--mig-worker-node-label-key=node-role.kubernetes.io/runai-mig-enabled",
				"--uninstall-detection-ref=run.ai/v1/Cluster//cluster",
			))
		})
	})

	Context("pod-group-assigner", func() {
		It("turns the Pod webhook off, since runai's admission-controller serves it", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner).NotTo(BeNil())
			Expect(podGroupAssigner.Webhooks).NotTo(BeNil())
			Expect(podGroupAssigner.Webhooks.EnablePodWebhook).To(Equal(ptr.To(false)))
		})

		It("carries the client rate on the service, not as extra args", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner.ExtraArgs).To(BeEmpty())
			Expect(podGroupAssigner.Service.K8sClientConfig).NotTo(BeNil())
			// KAI would otherwise default these to 20/100.
			Expect(podGroupAssigner.Service.K8sClientConfig.QPS).To(Equal(ptr.To(50)))
			Expect(podGroupAssigner.Service.K8sClientConfig.Burst).To(Equal(ptr.To(300)))
		})

		It("leaves image and disruption budget to KRM", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service).NotTo(BeNil())
			Expect(podGroupAssigner.Service.Image).To(BeNil())
			Expect(podGroupAssigner.Service.PodDisruptionBudget).To(BeNil())
		})

		It("carries the webhook ports and replica count", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner.ControllerService).NotTo(BeNil())
			Expect(podGroupAssigner.ControllerService.Webhook).NotTo(BeNil())
			Expect(podGroupAssigner.ControllerService.Webhook.Port).To(Equal(ptr.To(int32(443))))
			Expect(podGroupAssigner.ControllerService.Webhook.TargetPort).To(Equal(ptr.To(int32(8443))))
			Expect(podGroupAssigner.Replicas).NotTo(BeNil())
		})

		It("carries resources through", func() {
			runaiConfig := defaultRunaiConfig()
			podGroupAssigner := krmConfigFor(runaiConfig).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service.Resources).NotTo(BeNil())
		})

		It("preserves an overridden client rate", func() {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.PodGroupAssigner.K8sClientConfig.QPS = ptr.To(17)

			podGroupAssigner := krmConfigFor(runaiConfig).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service.K8sClientConfig.QPS).To(Equal(ptr.To(17)))
		})

		It("carries the enablement toggle through", func() {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.PodGroupAssigner.Enabled = ptr.To(false)

			podGroupAssigner := krmConfigFor(runaiConfig).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service.Enabled).To(Equal(ptr.To(false)))
		})
	})
})
