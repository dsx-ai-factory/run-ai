package resources

import (
	"context"
	"fmt"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

func ConfigMapForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		namespace = runaiConfig.Namespace
	)

	configMap := &v1.ConfigMap{}

	// Get the existing config map if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      deploymentName,
		Namespace: namespace,
	}, configMap)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	configMap.ObjectMeta.Name = deploymentName
	configMap.ObjectMeta.Namespace = namespace

	configMap.Data = webhooksFile(runaiConfig)

	return configMap, nil
}

func ConfigMapStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	configMap, err := ConfigMapForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if configMap.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("configmap [%s] is missing",
			configMap.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}

func webhooksFile(runaiConfig *runv1.RunaiConfig) map[string]string {
	var (
		namespace = runaiConfig.Namespace

		star = []string{"*"}
		v1   = []string{"v1"}

		mutatingWebhookTypeMeta = metav1.TypeMeta{
			APIVersion: admissionv1.GroupName + "/" + admissionv1.SchemeGroupVersion.Version,
			Kind:       "MutatingWebhookConfiguration",
		}

		validatingWebhookTypeMeta = metav1.TypeMeta{
			APIVersion: admissionv1.GroupName + "/" + admissionv1.SchemeGroupVersion.Version,
			Kind:       "ValidatingWebhookConfiguration",
		}

		webhookName                 = "runai-admission-controller.runai.svc"
		admissionServiceName        = "runai-admission-controller"
		anyworkloadWebhookName      = "anyworkload-controller.runai.svc"
		anyworkloadServiceName      = "anyworkload-controller"
		podGroupAssignerWebhookName = "pod-group-assigner.runai.svc"
		podGroupAssignerServiceName = "pod-group-assigner"
		accessRuleWebhookName       = "accessrule-controller.runai.svc"
		projectWebhookName          = "runai-project-controller.runai.svc"
		accessRuleServiceName       = operands.RunaiAccessRuleControllerName
		secretPvcWebhookName        = "runai-agent.runai.svc"
		secretPvcServiceName        = "runai-agent"
		applicationServiceName      = "application-controller"
		applicationWebhookName      = "application-validation.runai.svc"

		labels = map[string]string{
			"app": "runai",
		}

		runai           = []string{"run.ai"}
		schedulingrunai = []string{"scheduling.run.ai"}

		// TODO: use kai.resources after old org-unit crds cleanup
		runaiAndKairesources           = []string{"run.ai", "kai.resources"}
		schedulingrunaiAndKairesources = []string{"scheduling.run.ai", "kai.resources"}
		v1alpha1                       = []string{"v1alpha1"}
	)

	podMutatingWebhook := &admissionv1.MutatingWebhookConfiguration{
		TypeMeta: mutatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-pod",
			Labels: labels,
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name:                    webhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				NamespaceSelector:       utils.NamespaceSelectorRunaiQueue(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					admissionServiceName,
					"/mutate-pod"),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{admissionv1.Create},
						Rule: admissionv1.Rule{
							APIGroups:   star,
							APIVersions: star,
							Resources:   []string{"pods"},
							Scope:       scopeTypeNamespaced(),
						},
					},
				},
				ReinvocationPolicy: ptr.To(admissionv1.IfNeededReinvocationPolicy),
			},
		},
	}

	podControllerMutationMutatingWebhook := &admissionv1.MutatingWebhookConfiguration{
		TypeMeta: mutatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-pod-controller-mutation",
			Labels: labels,
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name:                    webhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				NamespaceSelector:       utils.NamespaceSelectorRunaiQueue(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					admissionServiceName,
					"/mutate-pod-controller"),
				ReinvocationPolicy: ptr.To(admissionv1.IfNeededReinvocationPolicy),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{admissionv1.Create, admissionv1.Update},
						Rule: admissionv1.Rule{
							APIGroups:   star,
							APIVersions: star,
							Resources:   podControllers(),
							Scope:       scopeTypeNamespaced(),
						},
					},
				},
			},
		},
	}

	podGroupMutationMutatingWebhook := &admissionv1.MutatingWebhookConfiguration{
		TypeMeta: mutatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-pod-group-mutation",
			Labels: labels,
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name:                    podGroupAssignerWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				NamespaceSelector:       utils.NamespaceSelectorRunaiQueue(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					podGroupAssignerServiceName,
					"/mutate-pod-group"),
				ReinvocationPolicy: ptr.To(admissionv1.IfNeededReinvocationPolicy),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{admissionv1.Create},
						Rule: admissionv1.Rule{
							APIGroups:   schedulingrunai,
							APIVersions: star,
							Resources:   []string{"podgroups"},
							Scope:       scopeTypeNamespaced(),
						},
					},
				},
			},
		},
	}

	anyworkloadRunnerValidationWebhook := &admissionv1.ValidatingWebhookConfiguration{
		TypeMeta: validatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-anyworkload-validation",
			Labels: labels,
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name:                    anyworkloadWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					anyworkloadServiceName,
					"/validate-workloadrunner"),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runai,
							APIVersions: v1alpha1,
							Resources:   []string{"workloadrunners"},
						},
					},
				},
			},
		},
	}

	anyworkloadRunnerMutationWebhook := &admissionv1.MutatingWebhookConfiguration{
		TypeMeta: mutatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-anyworkload-mutation",
			Labels: labels,
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name:                    anyworkloadWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					anyworkloadServiceName,
					"/mutate-workloadrunner"),
				ReinvocationPolicy: ptr.To(admissionv1.IfNeededReinvocationPolicy),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runai,
							APIVersions: v1alpha1,
							Resources:   []string{"workloadrunners"},
						},
					},
				},
			},
		},
	}
	ownershipRule := func(operations []admissionv1.OperationType, version string, resources []string) admissionv1.RuleWithOperations {
		return admissionv1.RuleWithOperations{
			Operations: operations,
			Rule: admissionv1.Rule{
				APIGroups: runai, APIVersions: []string{version}, Resources: resources,
				Scope: scopeTypeNamespaced(),
			},
		}
	}
	legacyOwnershipValidation := func(name string, rules []admissionv1.RuleWithOperations) *admissionv1.ValidatingWebhookConfiguration {
		return &admissionv1.ValidatingWebhookConfiguration{
			TypeMeta: validatingWebhookTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: name, Labels: labels,
			},
			Webhooks: []admissionv1.ValidatingWebhook{{
				Name: anyworkloadWebhookName, AdmissionReviewVersions: v1,
				SideEffects: sideEffectsNone(), FailurePolicy: failurePolicyTypeFail(),
				NamespaceSelector: utils.NamespaceSelectorRunaiQueue(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(
					namespace, anyworkloadServiceName, "/validate-workload-ownership-protection"),
				Rules: rules,
			}},
		}
	}
	legacyOwnershipMutation := func(name string, rules []admissionv1.RuleWithOperations) *admissionv1.MutatingWebhookConfiguration {
		return &admissionv1.MutatingWebhookConfiguration{
			TypeMeta: mutatingWebhookTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: name, Labels: labels,
			},
			Webhooks: []admissionv1.MutatingWebhook{{
				Name: anyworkloadWebhookName, AdmissionReviewVersions: v1,
				SideEffects: sideEffectsNone(), FailurePolicy: failurePolicyTypeFail(),
				NamespaceSelector: utils.NamespaceSelectorRunaiQueue(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(
					namespace, anyworkloadServiceName, "/mutate-workload-ownership"),
				ReinvocationPolicy: ptr.To(admissionv1.IfNeededReinvocationPolicy),
				Rules:              rules,
			}},
		}
	}
	var legacyOwnershipWebhooks []client.Object

	if *runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection {
		anyworkloadRunnerValidationWebhook.Webhooks = append(anyworkloadRunnerValidationWebhook.Webhooks,
			admissionv1.ValidatingWebhook{
				Name:                    "workloadrunner-ownership.anyworkload-controller.runai.svc",
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					anyworkloadServiceName,
					"/validate-workload-ownership-protection"),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
							admissionv1.Delete,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runai,
							APIVersions: v1alpha1,
							Resources:   []string{"workloadrunners"},
						},
					},
				},
			})
		anyworkloadRunnerMutationWebhook.Webhooks = append(anyworkloadRunnerMutationWebhook.Webhooks,
			admissionv1.MutatingWebhook{
				Name:                    "workloadrunner-ownership.anyworkload-controller.runai.svc",
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					anyworkloadServiceName,
					"/mutate-workload-ownership"),
				ReinvocationPolicy: ptr.To(admissionv1.IfNeededReinvocationPolicy),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Update,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runai,
							APIVersions: v1alpha1,
							Resources:   []string{"workloadrunners"},
						},
					},
				},
			})

		workloadValidationRules := []admissionv1.RuleWithOperations{ownershipRule(
			[]admissionv1.OperationType{admissionv1.Update, admissionv1.Delete},
			"v2alpha1", []string{"trainingworkloads", "interactiveworkloads", "distributedworkloads"},
		)}
		workloadMutationRules := []admissionv1.RuleWithOperations{ownershipRule(
			[]admissionv1.OperationType{admissionv1.Update},
			"v2alpha1", []string{"trainingworkloads", "interactiveworkloads", "distributedworkloads"},
		)}
		inferenceValidationRules := []admissionv1.RuleWithOperations{
			ownershipRule([]admissionv1.OperationType{admissionv1.Update, admissionv1.Delete},
				"v2alpha1", []string{"inferenceworkloads"}),
			ownershipRule([]admissionv1.OperationType{admissionv1.Update, admissionv1.Delete},
				"v1alpha1", []string{"distributedinferenceworkloads"}),
		}
		inferenceMutationRules := []admissionv1.RuleWithOperations{
			ownershipRule([]admissionv1.OperationType{admissionv1.Update},
				"v2alpha1", []string{"inferenceworkloads"}),
			ownershipRule([]admissionv1.OperationType{admissionv1.Update},
				"v1alpha1", []string{"distributedinferenceworkloads"}),
		}
		legacyOwnershipWebhooks = []client.Object{
			legacyOwnershipValidation("runai-workload-ownership-protection-validation", workloadValidationRules),
			legacyOwnershipValidation("runai-inference-workload-ownership-protection-validation", inferenceValidationRules),
			legacyOwnershipMutation("runai-workload-ownership-mutation", workloadMutationRules),
			legacyOwnershipMutation("runai-inference-workload-ownership-mutation", inferenceMutationRules),
		}
	}

	if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1internal.KubernetesFlavorOpenShift {
		ownershipWebhooks := append([]client.Object{
			anyworkloadRunnerValidationWebhook,
			anyworkloadRunnerMutationWebhook,
		}, legacyOwnershipWebhooks...)
		for _, webhook := range ownershipWebhooks {
			annotations := webhook.GetAnnotations()
			if annotations == nil {
				annotations = map[string]string{}
			}
			annotations["service.beta.openshift.io/inject-cabundle"] = "true"
			webhook.SetAnnotations(annotations)
		}
	}

	// The anyworkload-controller handler only allows WorkloadRunner workloadRefs, so this configuration is
	// exclusive to WorkloadRunner RestartRequests - RestartRequests for any other kind are denied at admission.
	restartRequestValidationWebhook := &admissionv1.ValidatingWebhookConfiguration{
		TypeMeta: validatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-restartrequest-validation",
			Labels: labels,
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name:                    anyworkloadWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				NamespaceSelector:       utils.NamespaceSelectorRunaiQueue(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					anyworkloadServiceName,
					"/validate-restartrequest"),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runai,
							APIVersions: v1alpha1,
							Resources:   []string{"restartrequests"},
							Scope:       scopeTypeNamespaced(),
						},
					},
				},
			},
		},
	}

	accessruleValidationWebhook := &admissionv1.ValidatingWebhookConfiguration{
		TypeMeta: validatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-accessrule-validation",
			Labels: labels,
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name:                    accessRuleWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					accessRuleServiceName,
					"/validate-access-rule"),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
							admissionv1.Delete,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runai,
							APIVersions: v1alpha1,
							Resources: []string{
								"accessrules",
							},
						},
					},
				},
			},
		},
	}

	projectValidationWebhook := &admissionv1.ValidatingWebhookConfiguration{
		TypeMeta: validatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-project-validation",
			Labels: labels,
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name:                    projectWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServicePathAndPort(namespace,
					"runai-agent",
					"/validate-project",
					9443),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
							admissionv1.Delete,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runaiAndKairesources,
							APIVersions: []string{"*"},
							Resources: []string{
								"projects",
							},
						},
					},
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
							admissionv1.Delete,
						},
						Rule: admissionv1.Rule{
							APIGroups:   schedulingrunaiAndKairesources,
							APIVersions: []string{"*"},
							Resources: []string{
								"departments",
							},
						},
					},
				},
			},
		},
	}

	secretPvcValidationWebhook := &admissionv1.ValidatingWebhookConfiguration{
		TypeMeta: validatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "runai-secret-pvc-validation",
			Labels: labels,
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name:                    secretPvcWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeIgnore(),
				ObjectSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      constants.RunaiAssetLabel,
							Operator: metav1.LabelSelectorOpExists,
						},
					},
				},
				// Webhook is served by the agent, on the same TLS port as
				// /validate-project. Kept hardcoded to 9443 to match the
				// sibling runai-project-validation entry above.
				ClientConfig: clientConfigWithNamespaceServicePathAndPort(namespace,
					secretPvcServiceName,
					"/validate-secrets-pvc",
					9443),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
							admissionv1.Delete,
						},
						Rule: admissionv1.Rule{
							APIGroups:   star,
							APIVersions: star,
							Resources:   []string{"secrets", "persistentvolumeclaims"},
							Scope:       scopeTypeNamespaced(),
						},
					},
				},
			},
		},
	}

	applicationValidationWebhook := &admissionv1.ValidatingWebhookConfiguration{
		TypeMeta: validatingWebhookTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   "application-validation",
			Labels: labels,
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name:                    applicationWebhookName,
				AdmissionReviewVersions: v1,
				SideEffects:             sideEffectsNone(),
				FailurePolicy:           failurePolicyTypeFail(),
				ClientConfig: clientConfigWithNamespaceServiceAndPath(namespace,
					applicationServiceName,
					"/validate-application"),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
							admissionv1.Delete,
						},
						Rule: admissionv1.Rule{
							APIGroups:   runai,
							APIVersions: star,
							Resources:   []string{"applications"},
							Scope:       scopeTypeNamespaced(),
						},
					},
				},
			},
		},
	}

	fileData := ptr.To[string]("")

	appendWebhookToFile(podMutatingWebhook, fileData)
	appendWebhookToFile(podControllerMutationMutatingWebhook, fileData)

	// anyworkload webhooks
	appendWebhookToFile(anyworkloadRunnerValidationWebhook, fileData)
	appendWebhookToFile(anyworkloadRunnerMutationWebhook, fileData)
	for _, webhook := range legacyOwnershipWebhooks {
		appendWebhookToFile(webhook, fileData)
	}

	appendWebhookToFile(restartRequestValidationWebhook, fileData)

	appendWebhookToFile(podGroupMutationMutatingWebhook, fileData)

	appendWebhookToFile(accessruleValidationWebhook, fileData)

	if !runaiConfig.Spec.Internal.Development.Headless {
		appendWebhookToFile(projectValidationWebhook, fileData)
	}

	appendWebhookToFile(secretPvcValidationWebhook, fileData)

	appendWebhookToFile(applicationValidationWebhook, fileData)

	return map[string]string{
		"webhooks.yaml": *fileData,
	}
}

func appendWebhookToFile(webhook client.Object, fileData *string) {
	*fileData += marshallOrPanic(webhook)
	*fileData += "---\n"
}

func marshallOrPanic(resource client.Object) string {
	yaml, err := yaml.Marshal(resource)
	if err != nil {
		panic(fmt.Errorf("failed marshaling resource [%s]: %v", resource.GetName(), err))
	}

	return string(yaml)
}

func sideEffectsNone() *admissionv1.SideEffectClass {
	none := admissionv1.SideEffectClassNone
	return &none
}

func scopeTypeNamespaced() *admissionv1.ScopeType {
	namespaced := admissionv1.NamespacedScope
	return &namespaced
}

func failurePolicyTypeFail() *admissionv1.FailurePolicyType {
	fail := admissionv1.Fail
	return &fail
}

func failurePolicyTypeIgnore() *admissionv1.FailurePolicyType {
	ignore := admissionv1.Ignore
	return &ignore
}

func serviceReferenceWithNamespaceServicePathAndPort(namespace, service, path string, port *int32) *admissionv1.ServiceReference {
	return &admissionv1.ServiceReference{
		Name:      service,
		Namespace: namespace,
		Path:      ptr.To[string](path),
		Port:      port,
	}
}

func clientConfigWithNamespaceServiceAndPath(namespace, service, path string) admissionv1.WebhookClientConfig {
	return admissionv1.WebhookClientConfig{
		Service: serviceReferenceWithNamespaceServicePathAndPort(namespace, service, path, nil),
	}
}

func clientConfigWithNamespaceServicePathAndPort(namespace, service, path string, port int32) admissionv1.WebhookClientConfig {
	return admissionv1.WebhookClientConfig{
		Service: serviceReferenceWithNamespaceServicePathAndPort(namespace, service, path, ptr.To(port)),
	}
}

func podControllers() []string {
	return []string{
		"statefulsets",
		"jobs",
		"deployments",
	}
}
