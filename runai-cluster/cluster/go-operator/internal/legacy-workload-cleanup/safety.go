/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"fmt"
	"slices"
	"strings"

	admissionv1 "k8s.io/api/admissionregistration/v1"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	legacyWorkloadFinalizer = "run.ai.workload/finalizer"

	workloadOwnershipValidationWebhook        = "runai-workload-ownership-protection-validation"
	inferenceOwnershipValidationWebhook       = "runai-inference-workload-ownership-protection-validation"
	workloadOwnershipMutationWebhook          = "runai-workload-ownership-mutation"
	inferenceWorkloadOwnershipMutationWebhook = "runai-inference-workload-ownership-mutation"
	anyWorkloadControllerService              = "anyworkload-controller"
	anyWorkloadControllerNamespace            = "runai"
)

func ownershipProtectionEnabled(config *runaiv1.RunaiConfig) bool {
	return config != nil && config.Spec.Global != nil &&
		config.Spec.Global.EnableWorkloadOwnershipProtection != nil &&
		*config.Spec.Global.EnableWorkloadOwnershipProtection
}

func withoutFinalizer(finalizers []string, allowlisted string) ([]string, bool) {
	remaining := make([]string, 0, len(finalizers))
	changed := false
	for _, finalizer := range finalizers {
		if finalizer == allowlisted {
			changed = true
			continue
		}
		remaining = append(remaining, finalizer)
	}
	return remaining, changed
}

var retiredResources = map[string]struct{}{
	"trainingworkloads": {}, "interactiveworkloads": {}, "distributedworkloads": {},
	"inferenceworkloads": {}, "distributedinferenceworkloads": {}, "runaijobs": {},
	"trainingpolicies": {}, "interactivepolicies": {}, "distributedpolicies": {},
	"inferencepolicies": {}, "distributedinferencepolicies": {},
}

var ownershipResourcesByConfiguration = map[string]map[string]string{
	workloadOwnershipValidationWebhook: {
		"trainingworkloads": "v2alpha1", "interactiveworkloads": "v2alpha1", "distributedworkloads": "v2alpha1",
	},
	inferenceOwnershipValidationWebhook: {
		"inferenceworkloads": "v2alpha1", "distributedinferenceworkloads": "v1alpha1",
	},
	workloadOwnershipMutationWebhook: {
		"trainingworkloads": "v2alpha1", "interactiveworkloads": "v2alpha1", "distributedworkloads": "v2alpha1",
	},
	inferenceWorkloadOwnershipMutationWebhook: {
		"inferenceworkloads": "v2alpha1", "distributedinferenceworkloads": "v1alpha1",
	},
}

func validateValidatingWebhooks(configurations []admissionv1.ValidatingWebhookConfiguration) error {
	allowed := map[string][]admissionv1.OperationType{
		workloadOwnershipValidationWebhook:  {admissionv1.Update, admissionv1.Delete},
		inferenceOwnershipValidationWebhook: {admissionv1.Update, admissionv1.Delete},
	}
	for _, configuration := range configurations {
		for _, webhook := range configuration.Webhooks {
			for _, rule := range webhook.Rules {
				if !targetsRetiredResource(rule.Rule) {
					continue
				}
				operations, approved := allowed[configuration.Name]
				if !approved {
					return fmt.Errorf("validating webhook configuration %s still targets a retired API", configuration.Name)
				}
				if err := validateOwnershipWebhookClient(configuration.Name, webhook.ClientConfig.Service); err != nil {
					return err
				}
				if !sameOperations(rule.Operations, operations) {
					return fmt.Errorf("validating webhook configuration %s operations %v, expected %v", configuration.Name, rule.Operations, operations)
				}
				if err := validateExactOwnershipRule(configuration.Name, rule.Rule); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateMutatingWebhooks(configurations []admissionv1.MutatingWebhookConfiguration) error {
	allowed := map[string][]admissionv1.OperationType{
		workloadOwnershipMutationWebhook:          {admissionv1.Update},
		inferenceWorkloadOwnershipMutationWebhook: {admissionv1.Update},
	}
	for _, configuration := range configurations {
		for _, webhook := range configuration.Webhooks {
			for _, rule := range webhook.Rules {
				if !targetsRetiredResource(rule.Rule) {
					continue
				}
				operations, approved := allowed[configuration.Name]
				if !approved {
					return fmt.Errorf("mutating webhook configuration %s still targets a retired API", configuration.Name)
				}
				if err := validateOwnershipWebhookClient(configuration.Name, webhook.ClientConfig.Service); err != nil {
					return err
				}
				if !sameOperations(rule.Operations, operations) {
					return fmt.Errorf("mutating webhook configuration %s operations %v, expected %v", configuration.Name, rule.Operations, operations)
				}
				if err := validateExactOwnershipRule(configuration.Name, rule.Rule); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateValidatingWebhooksForOwnershipMode(
	configurations []admissionv1.ValidatingWebhookConfiguration,
	ownershipEnabled bool,
) error {
	if ownershipEnabled {
		return validateValidatingWebhooks(configurations)
	}
	for _, configuration := range configurations {
		for _, webhook := range configuration.Webhooks {
			for _, rule := range webhook.Rules {
				if targetsRetiredResource(rule.Rule) {
					return fmt.Errorf("ownership protection is disabled but validating webhook configuration %s still targets a retired API",
						configuration.Name)
				}
			}
		}
	}
	return nil
}

func validateMutatingWebhooksForOwnershipMode(
	configurations []admissionv1.MutatingWebhookConfiguration,
	ownershipEnabled bool,
) error {
	if ownershipEnabled {
		return validateMutatingWebhooks(configurations)
	}
	for _, configuration := range configurations {
		for _, webhook := range configuration.Webhooks {
			for _, rule := range webhook.Rules {
				if targetsRetiredResource(rule.Rule) {
					return fmt.Errorf("ownership protection is disabled but mutating webhook configuration %s still targets a retired API",
						configuration.Name)
				}
			}
		}
	}
	return nil
}

func validateExactOwnershipRule(configurationName string, rule admissionv1.Rule) error {
	allowedResources := ownershipResourcesByConfiguration[configurationName]
	invalid := func() error {
		return fmt.Errorf("webhook configuration %s has a rule that is not an exact ownership handoff rule: groups=%v versions=%v resources=%v",
			configurationName, rule.APIGroups, rule.APIVersions, rule.Resources)
	}
	if len(rule.APIGroups) != 1 || rule.APIGroups[0] != "run.ai" || len(rule.APIVersions) != 1 || len(rule.Resources) == 0 {
		return invalid()
	}
	version := rule.APIVersions[0]
	for _, resource := range rule.Resources {
		expectedVersion, allowed := allowedResources[resource]
		if !allowed || version != expectedVersion {
			return invalid()
		}
	}
	return nil
}

func targetsRetiredResource(rule admissionv1.Rule) bool {
	if !slices.Contains(rule.APIGroups, "run.ai") && !slices.Contains(rule.APIGroups, "*") {
		return false
	}
	for _, resource := range rule.Resources {
		if resource == "*" {
			return true
		}
		baseResource, _, _ := strings.Cut(resource, "/")
		if baseResource == "*" {
			return true
		}
		if _, found := retiredResources[baseResource]; found {
			return true
		}
	}
	return false
}

func validateOwnershipWebhookClient(name string, service *admissionv1.ServiceReference) error {
	if service == nil || service.Name != anyWorkloadControllerService || service.Namespace != anyWorkloadControllerNamespace {
		return fmt.Errorf("ownership webhook configuration %s does not target %s/%s", name,
			anyWorkloadControllerNamespace, anyWorkloadControllerService)
	}
	return nil
}

func sameOperations(actual, expected []admissionv1.OperationType) bool {
	if len(actual) != len(expected) {
		return false
	}
	for _, operation := range expected {
		if !slices.Contains(actual, operation) {
			return false
		}
	}
	return true
}

func validateOwnershipHandoff(served []resourceDescriptor,
	validating []admissionv1.ValidatingWebhookConfiguration,
	mutating []admissionv1.MutatingWebhookConfiguration,
) error {
	for _, descriptor := range served {
		validatingName := ""
		mutatingName := ""
		switch descriptor.GVR.Resource {
		case "trainingworkloads", "interactiveworkloads", "distributedworkloads":
			validatingName = workloadOwnershipValidationWebhook
			mutatingName = workloadOwnershipMutationWebhook
		case "inferenceworkloads", "distributedinferenceworkloads":
			validatingName = inferenceOwnershipValidationWebhook
			mutatingName = inferenceWorkloadOwnershipMutationWebhook
		default:
			continue
		}
		if !validatingConfigurationCovers(validating, validatingName, descriptor) {
			return fmt.Errorf("required ownership handoff webhook configuration %s does not cover %s/%s %s",
				validatingName, descriptor.GVR.Group, descriptor.GVR.Version, descriptor.GVR.Resource)
		}
		if !mutatingConfigurationCovers(mutating, mutatingName, descriptor) {
			return fmt.Errorf("required ownership handoff webhook configuration %s does not cover %s/%s %s",
				mutatingName, descriptor.GVR.Group, descriptor.GVR.Version, descriptor.GVR.Resource)
		}
	}
	return nil
}

func validatingConfigurationCovers(configurations []admissionv1.ValidatingWebhookConfiguration,
	name string, descriptor resourceDescriptor,
) bool {
	for _, configuration := range configurations {
		if configuration.Name != name {
			continue
		}
		for _, webhook := range configuration.Webhooks {
			if validateOwnershipWebhookClient(name, webhook.ClientConfig.Service) != nil {
				continue
			}
			for _, rule := range webhook.Rules {
				if ruleCoversDescriptor(rule.Rule, descriptor) &&
					sameOperations(rule.Operations, []admissionv1.OperationType{admissionv1.Update, admissionv1.Delete}) {
					return true
				}
			}
		}
	}
	return false
}

func mutatingConfigurationCovers(configurations []admissionv1.MutatingWebhookConfiguration,
	name string, descriptor resourceDescriptor,
) bool {
	for _, configuration := range configurations {
		if configuration.Name != name {
			continue
		}
		for _, webhook := range configuration.Webhooks {
			if validateOwnershipWebhookClient(name, webhook.ClientConfig.Service) != nil {
				continue
			}
			for _, rule := range webhook.Rules {
				if ruleCoversDescriptor(rule.Rule, descriptor) &&
					sameOperations(rule.Operations, []admissionv1.OperationType{admissionv1.Update}) {
					return true
				}
			}
		}
	}
	return false
}

func ruleCoversDescriptor(rule admissionv1.Rule, descriptor resourceDescriptor) bool {
	return (slices.Contains(rule.APIGroups, descriptor.GVR.Group) || slices.Contains(rule.APIGroups, "*")) &&
		(slices.Contains(rule.APIVersions, descriptor.GVR.Version) || slices.Contains(rule.APIVersions, "*")) &&
		(slices.Contains(rule.Resources, descriptor.GVR.Resource) || slices.Contains(rule.Resources, "*"))
}
