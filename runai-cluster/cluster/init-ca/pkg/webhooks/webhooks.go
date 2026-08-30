package webhooks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/certificates"
	registrations "k8s.io/api/admissionregistration/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

const (
	deployVersionLabelName              = "deploy-version"
	podGroupAssignerWebhookName         = "runai-pod-group-mutation"    /* #nosec */
	accessruleControllerWebhookName     = "runai-accessrule-validation" /* #nosec */
	projectControllerWebhookName        = "runai-project-validation"    /* #nosec */
	secretPvcWebhookName                = "runai-secret-pvc-validation"
	applicationControllerWebhookName    = "application-validation"           /* #nosec */
	adviceRestartControllerWebhookName  = "runai-advice-restart-dw-mutation" /* #nosec */
	restartRequestValidationWebhookName = "runai-restartrequest-validation"  /* #nosec */
	workloadOwnershipValidationName     = "runai-workload-ownership-protection-validation"
	inferenceOwnershipValidationName    = "runai-inference-workload-ownership-protection-validation"
	workloadOwnershipMutationName       = "runai-workload-ownership-mutation"
	inferenceOwnershipMutationName      = "runai-inference-workload-ownership-mutation"
)

var workloadWebhookNamePattern = regexp.MustCompile(`(\w+)-(workload|policy)-(\w+)`)
var inferenceWorkloadWebhookNamePattern = regexp.MustCompile(`(\w+)-(inference-workload|inference-policy)-(\w+)`)
var anyworkloadControllerWebhookNamePattern = regexp.MustCompile(`(\w+)-(anyworkload)-(\w+)`)

type WebhookCertificates struct {
	AdmissionControllerCertificateDetails         *certificates.CertificateDetails
	InferenceWorkloadControllerCertificateDetails *certificates.CertificateDetails
	WorkloadControllerCertificateDetails          *certificates.CertificateDetails
	PodGroupAssignerCertificateDetails            *certificates.CertificateDetails
	AccessruleControllerCertificateDetails        *certificates.CertificateDetails
	AnyworkloadControllerCertificateDetails       *certificates.CertificateDetails
	ApplicationControllerCertificateDetails       *certificates.CertificateDetails
	AdviceRestartControllerCertificateDetails     *certificates.CertificateDetails
	AgentCertificateDetails                       *certificates.CertificateDetails
}

func ApplyWebhooks(
	ctx context.Context, client kubernetes.Interface,
	webhooksPath, deployVersion string,
	webhookCertificates *WebhookCertificates,
) error {
	log.Printf("Reading webhooks from %s", webhooksPath)
	validators, mutators, err := getWebhooksFromFile(
		webhooksPath, deployVersion,
		webhookCertificates,
	)
	if err != nil {
		return err
	}
	validators, mutators, err = filterOwnershipWebhooksByDiscovery(client.Discovery(), validators, mutators)
	if err != nil {
		return err
	}
	log.Printf("Applying webhooks")
	if err = createOrUpdateWebhooks(ctx, client, validators, mutators); err != nil {
		return err
	}
	log.Printf("Removing old webhooks")
	if err = deleteOldWebhooks(ctx, client, deployVersion); err != nil {
		return err
	}

	return nil
}

func filterOwnershipWebhooksByDiscovery(
	discoveryClient discovery.DiscoveryInterface,
	validators []*registrations.ValidatingWebhookConfiguration,
	mutators []*registrations.MutatingWebhookConfiguration,
) ([]*registrations.ValidatingWebhookConfiguration, []*registrations.MutatingWebhookConfiguration, error) {
	servedByGroupVersion := map[string]map[string]struct{}{}
	loadServedResources := func(groupVersion string) (map[string]struct{}, error) {
		if resources, loaded := servedByGroupVersion[groupVersion]; loaded {
			return resources, nil
		}
		resourceList, err := discoveryClient.ServerResourcesForGroupVersion(groupVersion)
		if err != nil {
			if k8serr.IsNotFound(err) {
				servedByGroupVersion[groupVersion] = map[string]struct{}{}
				return servedByGroupVersion[groupVersion], nil
			}
			return nil, err
		}
		resources := make(map[string]struct{}, len(resourceList.APIResources))
		for _, resource := range resourceList.APIResources {
			resources[resource.Name] = struct{}{}
		}
		servedByGroupVersion[groupVersion] = resources
		return resources, nil
	}

	filterRules := func(rules []registrations.RuleWithOperations) ([]registrations.RuleWithOperations, error) {
		filteredRules := make([]registrations.RuleWithOperations, 0, len(rules))
		for _, rule := range rules {
			filteredResources := make([]string, 0, len(rule.Resources))
			for _, resource := range rule.Resources {
				served := false
				for _, group := range rule.APIGroups {
					for _, version := range rule.APIVersions {
						resources, err := loadServedResources(group + "/" + version)
						if err != nil {
							return nil, err
						}
						if _, exists := resources[resource]; exists {
							served = true
						}
					}
				}
				if served {
					filteredResources = append(filteredResources, resource)
				}
			}
			if len(filteredResources) != 0 {
				rule.Resources = filteredResources
				filteredRules = append(filteredRules, rule)
			}
		}
		return filteredRules, nil
	}

	filteredValidators := make([]*registrations.ValidatingWebhookConfiguration, 0, len(validators))
	for _, validator := range validators {
		if validator.Name != workloadOwnershipValidationName && validator.Name != inferenceOwnershipValidationName {
			filteredValidators = append(filteredValidators, validator)
			continue
		}
		filtered := validator.DeepCopy()
		filtered.Webhooks = nil
		for _, webhook := range validator.Webhooks {
			rules, err := filterRules(webhook.Rules)
			if err != nil {
				return nil, nil, err
			}
			if len(rules) != 0 {
				webhook.Rules = rules
				filtered.Webhooks = append(filtered.Webhooks, webhook)
			}
		}
		if len(filtered.Webhooks) != 0 {
			filteredValidators = append(filteredValidators, filtered)
		}
	}

	filteredMutators := make([]*registrations.MutatingWebhookConfiguration, 0, len(mutators))
	for _, mutator := range mutators {
		if mutator.Name != workloadOwnershipMutationName && mutator.Name != inferenceOwnershipMutationName {
			filteredMutators = append(filteredMutators, mutator)
			continue
		}
		filtered := mutator.DeepCopy()
		filtered.Webhooks = nil
		for _, webhook := range mutator.Webhooks {
			rules, err := filterRules(webhook.Rules)
			if err != nil {
				return nil, nil, err
			}
			if len(rules) != 0 {
				webhook.Rules = rules
				filtered.Webhooks = append(filtered.Webhooks, webhook)
			}
		}
		if len(filtered.Webhooks) != 0 {
			filteredMutators = append(filteredMutators, filtered)
		}
	}

	return filteredValidators, filteredMutators, nil
}

func setWebhookMetadata(
	deployVersion string,
	objectMeta *metav1.ObjectMeta,
) {
	objectMeta.Labels[deployVersionLabelName] = deployVersion
}

func setClientConfigCA(caBundle []byte, client *registrations.WebhookClientConfig) {
	if len(caBundle) > 0 {
		client.CABundle = caBundle
	}
}

func setVWCData(
	vwc *registrations.ValidatingWebhookConfiguration,
	deployVersion string, cd *certificates.CertificateDetails,
) {
	setWebhookMetadata(deployVersion, &vwc.ObjectMeta)
	for ind := range vwc.Webhooks {
		setClientConfigCA(cd.CaBundle, &vwc.Webhooks[ind].ClientConfig)
	}
}

func setMWCData(
	mwc *registrations.MutatingWebhookConfiguration,
	deployVersion string, cd *certificates.CertificateDetails,
) {
	setWebhookMetadata(deployVersion, &mwc.ObjectMeta)
	for ind := range mwc.Webhooks {
		setClientConfigCA(cd.CaBundle, &mwc.Webhooks[ind].ClientConfig)
	}
}

// read kubernetes resources from file
func readKubernetesResources(path string) (objects []unstructured.Unstructured, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		objects = append(objects, *obj)
	}
	return objects, nil
}

// cast all unstructured.Unstructured objects to validatingwebhookconfiguration or mutatingwebhookconfiguration
func castToWebhookConfigurations(objects []unstructured.Unstructured) ([]*registrations.ValidatingWebhookConfiguration, []*registrations.MutatingWebhookConfiguration) {
	var validators []*registrations.ValidatingWebhookConfiguration
	var mutators []*registrations.MutatingWebhookConfiguration
	for _, obj := range objects {
		kind := obj.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case "ValidatingWebhookConfiguration":
			var validator registrations.ValidatingWebhookConfiguration
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &validator); err != nil {
				log.Printf("Failed to convert unstructured object to validatingwebhookconfiguration: %v", err)
				continue
			}
			validators = append(validators, &validator)
		case "MutatingWebhookConfiguration":
			var mutator registrations.MutatingWebhookConfiguration
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &mutator); err != nil {
				log.Printf("Failed to convert unstructured object to mutatingwebhookconfiguration: %v", err)
				continue
			}
			mutators = append(mutators, &mutator)
		default:
			log.Printf("Unknown webhook type: %s", obj.GetObjectKind().GroupVersionKind().Kind)
		}
	}
	return validators, mutators
}

func chooseCertificateDetails(name string, webhookCertificates *WebhookCertificates) *certificates.CertificateDetails {
	switch name {
	case workloadOwnershipValidationName, inferenceOwnershipValidationName,
		workloadOwnershipMutationName, inferenceOwnershipMutationName:
		return webhookCertificates.AnyworkloadControllerCertificateDetails
	}
	//
	//   naming of workload-controller webhook
	//     runai-training-policy-validation
	//           ---a---- --b--
	//   where
	//     (a) is the resource type: training, interactive, deployment, inference
	//     (b) is either workload or policy
	//
	if inferenceWorkloadWebhookNamePattern.FindString(name) != "" {
		return webhookCertificates.InferenceWorkloadControllerCertificateDetails
	}
	if workloadWebhookNamePattern.FindString(name) != "" {
		return webhookCertificates.WorkloadControllerCertificateDetails
	}
	if anyworkloadControllerWebhookNamePattern.FindString(name) != "" {
		return webhookCertificates.AnyworkloadControllerCertificateDetails
	}
	if name == podGroupAssignerWebhookName {
		return webhookCertificates.PodGroupAssignerCertificateDetails
	}
	if name == accessruleControllerWebhookName {
		return webhookCertificates.AccessruleControllerCertificateDetails
	}
	if name == projectControllerWebhookName {
		// The project/department validating webhook is served by the agent, so its
		// CA bundle must match the agent's serving certificate.
		return webhookCertificates.AgentCertificateDetails
	}
	if name == secretPvcWebhookName {
		return webhookCertificates.AgentCertificateDetails
	}
	if name == applicationControllerWebhookName {
		return webhookCertificates.ApplicationControllerCertificateDetails
	}
	if name == adviceRestartControllerWebhookName {
		return webhookCertificates.AdviceRestartControllerCertificateDetails
	}
	if name == restartRequestValidationWebhookName {
		// served by the anyworkload-controller, so its CA bundle must match that serving certificate
		return webhookCertificates.AnyworkloadControllerCertificateDetails
	}

	return webhookCertificates.AdmissionControllerCertificateDetails
}

func getWebhooksFromFile(
	webhooksPath, deployVersion string,
	webhookCertificates *WebhookCertificates,
) (
	validators []*registrations.ValidatingWebhookConfiguration,
	mutators []*registrations.MutatingWebhookConfiguration,
	err error,
) {
	objects, err := readKubernetesResources(webhooksPath)
	if err != nil {
		log.Printf("Failed to read webhooks from %s: %v", webhooksPath, err)
		return nil, nil, err
	}
	validators, mutators = castToWebhookConfigurations(objects)

	filteredValidators := make([]*registrations.ValidatingWebhookConfiguration, 0, len(validators))
	for _, vwc := range validators {
		certificateDetails := chooseCertificateDetails(vwc.Name, webhookCertificates)
		if certificateDetails == nil {
			log.Printf("Skipping validating webhook %s because its serving certificate is disabled", vwc.Name)
			continue
		}
		setVWCData(vwc, deployVersion, certificateDetails)
		filteredValidators = append(filteredValidators, vwc)
	}
	validators = filteredValidators

	filteredMutators := make([]*registrations.MutatingWebhookConfiguration, 0, len(mutators))
	for _, mwc := range mutators {
		certificateDetails := chooseCertificateDetails(mwc.Name, webhookCertificates)
		if certificateDetails == nil {
			log.Printf("Skipping mutating webhook %s because its serving certificate is disabled", mwc.Name)
			continue
		}
		setMWCData(mwc, deployVersion, certificateDetails)
		filteredMutators = append(filteredMutators, mwc)
	}
	mutators = filteredMutators
	return
}

func updateVWCFromExisting(
	ctx context.Context, client kubernetes.Interface,
	existingVWC *registrations.ValidatingWebhookConfiguration, vwc *registrations.ValidatingWebhookConfiguration,
) error {
	existingVWC.Webhooks = vwc.Webhooks
	mergeDesiredAnnotations(existingVWC, vwc)
	setVWCData(existingVWC, vwc.Labels[deployVersionLabelName],
		certificates.NewCertificateDetails(vwc.Webhooks[0].ClientConfig.CABundle))
	if _, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(ctx, existingVWC, metav1.UpdateOptions{}); err != nil {
		log.Printf("Failed to update validating webhook: %s\n%v", vwc.Name, err)
		return err
	}
	return nil
}

func updateMWCFromExisting(
	ctx context.Context, client kubernetes.Interface,
	existingMWC *registrations.MutatingWebhookConfiguration, mwc *registrations.MutatingWebhookConfiguration,
) error {
	existingMWC.Webhooks = mwc.Webhooks
	mergeDesiredAnnotations(existingMWC, mwc)
	setMWCData(existingMWC, mwc.Labels[deployVersionLabelName],
		certificates.NewCertificateDetails(mwc.Webhooks[0].ClientConfig.CABundle))
	if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, existingMWC, metav1.UpdateOptions{}); err != nil {
		log.Printf("Failed to update mutating webhook: %s\n%v", mwc.Name, err)
		return err
	}
	return nil
}

func mergeDesiredAnnotations(existing, desired metav1.Object) {
	annotations := existing.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	for key, value := range desired.GetAnnotations() {
		annotations[key] = value
	}
	existing.SetAnnotations(annotations)
}

func createOrUpdateWebhooks(
	ctx context.Context, client kubernetes.Interface,
	validators []*registrations.ValidatingWebhookConfiguration, mutators []*registrations.MutatingWebhookConfiguration,
) error {
	for _, validator := range validators {
		existingVWC, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(ctx, validator.Name, metav1.GetOptions{})
		if k8serr.IsNotFound(err) {
			log.Printf("Validating webhook %s does not exist, creating it", validator.Name)
			if _, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(ctx, validator, metav1.CreateOptions{}); err != nil {
				log.Printf("Failed to create validating webhook: %s\n%v", validator.Name, err)
				return err
			}
		} else if err != nil {
			log.Printf("Failed to get validating webhook: %s\n%v", validator.Name, err)
			return err
		} else {
			log.Printf("Validating webhook %s exists, updating it", validator.Name)
			if err := updateVWCFromExisting(ctx, client, existingVWC, validator); err != nil {
				return err
			}
		}
	}

	for _, mutator := range mutators {
		existingMWC, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mutator.Name, metav1.GetOptions{})
		if k8serr.IsNotFound(err) {
			log.Printf("Mutating webhook %s does not exist, creating it", mutator.Name)
			if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, mutator, metav1.CreateOptions{}); err != nil {
				log.Printf("Failed to create mutating webhook: %s\n%v", mutator.Name, err)
				return err
			}
		} else if err != nil {
			log.Printf("Failed to get mutating webhook: %s\n%v", mutator.Name, err)
			return err
		} else {
			log.Printf("Mutating webhook %s exists, updating it", mutator.Name)
			if err := updateMWCFromExisting(ctx, client, existingMWC, mutator); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteOldWebhooks(ctx context.Context, client kubernetes.Interface, deployVersion string) error {
	oldWebhooks, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=runai,deploy-version!=%s", deployVersion),
	})
	if err != nil {
		log.Printf("Failed to list and delete old webhooks: %v", err)
		return err
	}

	for _, webhook := range oldWebhooks.Items {
		log.Printf("Deleting old webhook: %s", webhook.Name)
		err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, webhook.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Failed to delete old webhook: %s\n%v", webhook.Name, err)
			return err
		}
	}

	oldWebhooksV, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=runai,deploy-version!=%s", deployVersion),
	})
	if err != nil {
		log.Printf("Failed to list and delete old webhooks: %v", err)
		return err
	}

	for _, webhook := range oldWebhooksV.Items {
		log.Printf("Deleting old webhook: %s", webhook.Name)
		err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(ctx, webhook.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Failed to delete old webhook: %s\n%v", webhook.Name, err)
			return err
		}
	}

	return nil
}
