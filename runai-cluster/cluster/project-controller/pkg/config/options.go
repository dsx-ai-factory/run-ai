/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"flag"
	"os"

	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	kaipgconstants "github.com/kai-scheduler/KAI-scheduler/pkg/podgrouper/podgrouper/plugins/constants"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

type Options struct {
	MetricsPort          string
	Debug                bool
	EnableLeaderElection bool
}

func SetOptions() (*Options, *ProjectReconcilerConfig, error) {
	return parse(flag.CommandLine, os.Args[1:])
}

func parse(flagSet *flag.FlagSet, options []string) (*Options, *ProjectReconcilerConfig, error) {
	opts := &Options{}
	projectReconcilerConfig := &ProjectReconcilerConfig{}
	if err := parseCommandLineArgs(flagSet, options, opts, projectReconcilerConfig); err != nil {
		return nil, nil, err
	}
	// Make the parsed config available to deep-stack handlers/reconcilers via Get().
	current = projectReconcilerConfig
	return opts, projectReconcilerConfig, nil
}

func parseCommandLineArgs(flagSet *flag.FlagSet, options []string, opts *Options, config *ProjectReconcilerConfig) error {
	flagSet.StringVar(&opts.MetricsPort, "metrics-port", "9400", "The port to serve metrics from")
	flagSet.BoolVar(&opts.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flagSet.BoolVar(&config.CreateNamespaces, "namespaces", false, "Create namespaces for a project")
	flagSet.BoolVar(&config.CreateRoleBindings, "role-bindings", false, "Create RoleBindings for a project based on --rolebindings-configmap-name configmap in --rolebindings-configmap-namespace namespace")
	flagSet.BoolVar(&opts.Debug, "debug", false, "Should use debug log level")
	flagSet.StringVar(&config.RoleBindingsCm, "rolebindings-configmap-name", "", "The cm to read role bindings from")
	flagSet.StringVar(&config.RoleBindingsCmNamespace, "rolebindings-configmap-namespace", "", "The namespace of the cm to read role bindings from")
	flagSet.BoolVar(&config.ClusterWideSecrets, "cluster-wide-secrets", false,
		"Enable cluster wide secret feature, means the project controller will listen on all secrets and create secrets")
	flagSet.BoolVar(&config.ClusterWidePvcs, "cluster-wide-pvcs", false,
		"Enable cluster wide pvcs feature, means the project controller will listen on all pvcs and create pvcs")
	flagSet.BoolVar(&config.ClusterWideConfigMaps, "cluster-wide-config-maps", true,
		"Enable cluster wide config maps feature, means the project controller will listen on all config maps and create config maps")
	flagSet.BoolVar(&config.LimitRange, "limit-range", false, "Enable creation of runai limit range")
	flagSet.BoolVar(&config.IsOpenshift, "openshift", false, "Enable openshift specific features")

	flagSet.StringVar(&config.LimitRangeName, "limit-range-name", defaultLimitRangeName, "Name of the LimitRange resource created in each project namespace")
	flagSet.StringVar(&config.GvkDeleteBlockersNamespace, "project-delete-blockers-namespace", "",
		"Namespace of the 'project-delete-blockers' ConfigMap describing project-deletion blockers. Empty (default) means no blockers - deletion is never blocked")
	// Label-vocabulary flags — KAI-style defaults; go-operator's deployments.go
	// passes runai overrides so runai deployment behavior is unchanged.
	flagSet.StringVar(&config.NodePoolLabelKey, "nodepool-label-key", kaiconstants.DefaultNodePoolLabelKey, "Label key used to associate Kubernetes resources with a node pool")
	flagSet.StringVar(&config.DefaultNodepoolName, "default-nodepool-name", kaiconstants.DefaultNodePoolName, "Name of the implicit default node pool (its queues carry no node-pool label)")
	flagSet.StringVar(&config.QueueLabelKey, "queue-label-key", kaiconstants.DefaultQueueLabel, "Label key used to identify the queue of a workload")
	flagSet.StringVar(&config.NamespaceProjectLabelKey, "namespace-project-label-key", kaiv1alpha1.NamespaceProjectLabelKey, "Label key written on a project's namespace to identify the project it belongs to")
	flagSet.StringVar(&config.ProjectLabelKey, "project-label-key", kaipgconstants.ProjectLabelKey, "Label key used to identify the project of a workload")
	flagSet.StringVar(&config.ProjectIdLabelKey, "project-id-label-key", defaultProjectIdLabelKey, "Label key used to identify the numeric project ID on resources")
	flagSet.StringVar(&config.QueueDepartmentNameLabelKey, "queue-department-name-label-key", defaultQueueDepartmentNameLabelKey, "Label key written on queues to record the department they belong to")
	flagSet.StringVar(&config.NamespaceVersionLabelKey, "namespace-version-label-key", defaultNamespaceVersionLabelKey, "Label key written on managed namespaces to track their schema version")
	flagSet.StringVar(&config.EnforceSchedulerAnnotationKey, "enforce-scheduler-annotation-key", defaultEnforceSchedulerAnnotationKey,
		"Annotation key written on a project's namespace carrying Project.Spec.EnforceKaiScheduler; read back by the component that enforces the scheduler on pods")
	flagSet.StringVar(&config.FinalizerDomain, "finalizer-domain", defaultFinalizerDomain, "DNS-style domain prefix used to compose the project finalizer string")
	flagSet.StringVar(&config.InstallNamespace, "install-namespace", "",
		"Namespace project-controller is installed in; where it reads replicated-resource sources"+
			" (limit-range ConfigMap, cluster-wide secrets) from and watches")
	flagSet.StringVar(&config.ProjectNamePrefix, "project-name-prefix", defaultProjectNamePrefix,
		"Prefix used to name and identify the namespaces project-controller manages (<prefix>-<project>)")
	flagSet.StringVar(&config.ResourceManualOverrideLabelKey, "resource-manual-override-label-key", defaultResourceManualOverrideLabelKey,
		"Label key that, when set to 'true' on a managed resource, marks it as manually overridden so project-controller will not create, update, or delete it")
	flagSet.BoolVar(&config.EnableProjectValidationWebhook, "enable-project-validation-webhook", defaultEnableProjectValidationWebhook,
		"Register the validating admission webhook handler for Project resources")
	flagSet.BoolVar(&config.EnableDepartmentValidationWebhook, "enable-department-validation-webhook", defaultEnableDepartmentValidationWebhook,
		"Register the validating admission webhook handler for Department resources")
	flagSet.IntVar(&config.WebhookPort, "webhook-port", 8443, "Port the validating webhook server listens on")
	flagSet.StringVar(&config.WebhookCertDir, "cert-dir", "/etc/webhook/certs",
		"Directory that contains the webhook server's TLS key and certificate")
	flagSet.BoolVar(&config.EnableProfiler, "enable-profiling", false, "Enable profiling")
	flagSet.StringVar(&config.ProfilerApiPort, "profiler-api-port", "8182", "Profiler API port")
	flagSet.IntVar(&config.K8sClientConfigQPS, "qps", 50, "Queries per second to the K8s API server")
	flagSet.IntVar(&config.K8sClientConfigBurst, "burst", 300, "Burst to the K8s API server")
	return flagSet.Parse(options)
}
