/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"fmt"
	"time"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ownershipValidationTimeout bounds authorization reads. Callers must still
// supply a live, uncached API reader; the timeout is only defense in depth.
const ownershipValidationTimeout = 30 * time.Second

const (
	podKind     = "Pod"
	serviceKind = "Service"
)

var (
	// errOwnershipViolation marks a live graph that is unsafe or cannot prove the
	// expected WorkloadRunner ownership. Other errors remain retryable API failures.
	errOwnershipViolation           = errors.New("networking ownership violation")
	errOwnershipReaderNotConfigured = errors.New("ownership reader is not configured")
)

// validateDesiredServiceOwnership checks the selector of a plugin-created
// Service before it is applied. An empty target set is safe: it routes nowhere.
func validateDesiredServiceOwnership(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	desired *corev1.Service,
) error {
	ctx, cancel := context.WithTimeout(ctx, ownershipValidationTimeout)
	defer cancel()

	if reader == nil {
		return errOwnershipReaderNotConfigured
	}
	validationReader := newOwnershipValidationReader(reader)
	if desired == nil {
		return ownershipViolationf("desired Service is nil")
	}
	if err := validateRunnerAndServiceKey(
		runner,
		types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name},
	); err != nil {
		return err
	}
	if err := validateServiceShape(runner, desired); err != nil {
		return err
	}
	return validateSelectorPods(
		ctx, validationReader, runner, desired, make(map[endpointPodIdentity]*corev1.Pod),
	)
}

// validateCreatedServiceOwnership proves both the plugin-created Service owner
// and every current traffic destination.
func validateCreatedServiceOwnership(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	key types.NamespacedName,
) error {
	ctx, cancel := context.WithTimeout(ctx, ownershipValidationTimeout)
	defer cancel()

	return validateServiceOwnership(ctx, reader, runner, key, true)
}

// validateReferencedServiceOwnership proves every current traffic destination
// without requiring, adopting, or mutating the referenced Service itself.
func validateReferencedServiceOwnership(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	key types.NamespacedName,
) error {
	ctx, cancel := context.WithTimeout(ctx, ownershipValidationTimeout)
	defer cancel()

	return validateServiceOwnership(ctx, reader, runner, key, false)
}

func validateServiceOwnership(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	key types.NamespacedName,
	requireServiceOwner bool,
) error {
	if reader == nil {
		return errOwnershipReaderNotConfigured
	}
	validationReader := newOwnershipValidationReader(reader)
	if err := validateRunnerAndServiceKey(runner, key); err != nil {
		return err
	}

	service := &corev1.Service{}
	if err := validationReader.Get(ctx, key, service); err != nil {
		return fmt.Errorf("read Service %s: %w", key, err)
	}
	if service.UID == "" {
		return ownershipViolationf("Service %s has no live UID", key)
	}
	if err := validateServiceShape(runner, service); err != nil {
		return err
	}
	verifiedPods := make(map[endpointPodIdentity]*corev1.Pod)
	var errs []error
	if requireServiceOwner {
		errs = append(errs, validateObjectOwner(ctx, validationReader, runner, service, "Service"))
	}
	errs = append(errs,
		validateSelectorPods(ctx, validationReader, runner, service, verifiedPods),
		validateEndpointSlices(ctx, validationReader, runner, service, verifiedPods),
	)
	return prioritizeOwnershipErrors(errs...)
}

func validateRunnerAndServiceKey(
	runner *dromosv1alpha1.WorkloadRunner,
	key types.NamespacedName,
) error {
	if err := validateRunnerIdentityIsComplete(runner); err != nil {
		return err
	}
	if key.Namespace != runner.Namespace {
		return ownershipViolationf(
			"Service namespace %q does not match WorkloadRunner namespace %q",
			key.Namespace, runner.Namespace,
		)
	}
	if key.Name == "" {
		return ownershipViolationf("Service name is empty")
	}
	return nil
}

func validateRunnerIdentityIsComplete(runner *dromosv1alpha1.WorkloadRunner) error {
	if runner == nil {
		return ownershipViolationf("WorkloadRunner is nil")
	}
	if runner.Namespace == "" || runner.Name == "" || runner.UID == "" {
		return ownershipViolationf("WorkloadRunner identity is incomplete")
	}
	return nil
}

func validateServiceShape(
	runner *dromosv1alpha1.WorkloadRunner,
	service *corev1.Service,
) error {
	if service.Namespace != runner.Namespace {
		return ownershipViolationf(
			"Service namespace %q does not match WorkloadRunner namespace %q",
			service.Namespace, runner.Namespace,
		)
	}
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return ownershipViolationf("ExternalName Service targets cannot be proven as workload Pods")
	}
	if len(service.Spec.Selector) == 0 {
		return ownershipViolationf("selectorless Service targets cannot be proven as workload Pods")
	}
	return nil
}

func validateSelectorPods(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	service *corev1.Service,
	verifiedPods map[endpointPodIdentity]*corev1.Pod,
) error {
	var pods corev1.PodList
	if err := reader.List(
		ctx,
		&pods,
		client.InNamespace(service.Namespace),
		client.MatchingLabels(service.Spec.Selector),
	); err != nil {
		return fmt.Errorf("list Pods selected by Service %s/%s: %w", service.Namespace, service.Name, err)
	}

	errs := make([]error, 0, len(pods.Items))
	for i := range pods.Items {
		pod := &pods.Items[i]
		err := validateObjectOwner(ctx, reader, runner, pod, "selected Pod")
		errs = append(errs, err)
		if err == nil {
			verifiedPods[endpointPodIdentity{Name: pod.Name, UID: pod.UID}] = pod
		}
	}
	return prioritizeOwnershipErrors(errs...)
}

func validateEndpointSlices(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	service *corev1.Service,
	verifiedPods map[endpointPodIdentity]*corev1.Pod,
) error {
	var slices discoveryv1.EndpointSliceList
	if err := reader.List(
		ctx,
		&slices,
		client.InNamespace(service.Namespace),
		client.MatchingLabels{discoveryv1.LabelServiceName: service.Name},
	); err != nil {
		return fmt.Errorf("list EndpointSlices for Service %s/%s: %w", service.Namespace, service.Name, err)
	}

	errs := make([]error, 0)
	for i := range slices.Items {
		slice := &slices.Items[i]
		sliceErrs := []error{
			validateEndpointSliceService(service, slice),
			validateEndpointSliceAddressType(slice),
		}
		if err := prioritizeOwnershipErrors(sliceErrs...); err != nil {
			errs = append(errs, err)
			continue
		}
		for endpointIndex := range slice.Endpoints {
			endpoint := &slice.Endpoints[endpointIndex]
			pod, err := validateEndpointPod(ctx, reader, runner, service, slice, endpoint, verifiedPods)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if err := validateEndpointAddresses(slice, endpoint, pod); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return prioritizeOwnershipErrors(errs...)
}

func validateEndpointSliceAddressType(slice *discoveryv1.EndpointSlice) error {
	switch slice.AddressType {
	case discoveryv1.AddressTypeIPv4, discoveryv1.AddressTypeIPv6:
		return nil
	default:
		return ownershipViolationf(
			"EndpointSlice %s/%s has unsupported address type %q",
			slice.Namespace, slice.Name, slice.AddressType,
		)
	}
}

func validateEndpointSliceService(
	service *corev1.Service,
	slice *discoveryv1.EndpointSlice,
) error {
	controller := metav1.GetControllerOfNoCopy(slice)
	if controller == nil {
		return ownershipViolationf(
			"EndpointSlice %s/%s has no controller Service owner", slice.Namespace, slice.Name,
		)
	}
	if controller.APIVersion != corev1.SchemeGroupVersion.String() || controller.Kind != serviceKind ||
		controller.Name != service.Name || controller.UID != service.UID {
		return ownershipViolationf(
			"EndpointSlice %s/%s owner does not match Service %s/%s UID %q",
			slice.Namespace, slice.Name, service.Namespace, service.Name, service.UID,
		)
	}
	return nil
}

type endpointPodIdentity struct {
	Name string
	UID  types.UID
}

func validateEndpointPod(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	service *corev1.Service,
	slice *discoveryv1.EndpointSlice,
	endpoint *discoveryv1.Endpoint,
	verifiedPods map[endpointPodIdentity]*corev1.Pod,
) (*corev1.Pod, error) {
	ref := endpoint.TargetRef
	if ref == nil {
		return nil, ownershipViolationf(
			"EndpointSlice %s/%s contains an IP-only endpoint", slice.Namespace, slice.Name,
		)
	}
	if (ref.APIVersion != "" && ref.APIVersion != corev1.SchemeGroupVersion.String()) || ref.Kind != podKind ||
		ref.Namespace != service.Namespace || ref.Name == "" || ref.UID == "" {
		return nil, ownershipViolationf(
			"EndpointSlice %s/%s contains an invalid Pod target reference", slice.Namespace, slice.Name,
		)
	}

	identity := endpointPodIdentity{Name: ref.Name, UID: ref.UID}
	if pod, found := verifiedPods[identity]; found {
		return pod, nil
	}

	pod := &corev1.Pod{}
	key := types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}
	if err := reader.Get(ctx, key, pod); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ownershipViolationf(
				"EndpointSlice %s/%s targets missing Pod %s", slice.Namespace, slice.Name, key,
			)
		}
		return nil, fmt.Errorf("read endpoint Pod %s: %w", key, err)
	}
	if pod.UID != ref.UID {
		return nil, ownershipViolationf(
			"EndpointSlice %s/%s targets stale Pod UID %q; live Pod %s has UID %q",
			slice.Namespace, slice.Name, ref.UID, key, pod.UID,
		)
	}
	if err := validateObjectOwner(ctx, reader, runner, pod, "endpoint Pod"); err != nil {
		return nil, err
	}
	verifiedPods[identity] = pod
	return pod, nil
}

func validateEndpointAddresses(
	slice *discoveryv1.EndpointSlice,
	endpoint *discoveryv1.Endpoint,
	pod *corev1.Pod,
) error {
	if len(endpoint.Addresses) == 0 {
		return ownershipViolationf(
			"EndpointSlice %s/%s contains an endpoint without addresses", slice.Namespace, slice.Name,
		)
	}

	podIPs := make(map[string]struct{}, len(pod.Status.PodIPs)+1)
	if pod.Status.PodIP != "" {
		podIPs[pod.Status.PodIP] = struct{}{}
	}
	for _, podIP := range pod.Status.PodIPs {
		if podIP.IP != "" {
			podIPs[podIP.IP] = struct{}{}
		}
	}
	for _, address := range endpoint.Addresses {
		if _, found := podIPs[address]; !found {
			return ownershipViolationf(
				"EndpointSlice %s/%s address %q does not belong to Pod %s/%s",
				slice.Namespace, slice.Name, address, pod.Namespace, pod.Name,
			)
		}
	}
	return nil
}

// ownershipValidationReader memoizes successful live reads of intermediate
// owner metadata only for the duration of one validation call. It prevents
// Pods that share a controller chain from repeatedly hitting the API server;
// no result survives into a later reconciliation or authorization decision.
type ownershipValidationReader struct {
	client.Reader
	owners map[ownershipValidationOwnerKey]*metav1.PartialObjectMetadata
}

type ownershipValidationOwnerKey struct {
	key client.ObjectKey
	gvk schema.GroupVersionKind
}

func newOwnershipValidationReader(reader client.Reader) client.Reader {
	if _, ok := reader.(*ownershipValidationReader); ok {
		return reader
	}
	return &ownershipValidationReader{
		Reader: reader,
		owners: make(map[ownershipValidationOwnerKey]*metav1.PartialObjectMetadata),
	}
}

func (r *ownershipValidationReader) Get(
	ctx context.Context,
	key client.ObjectKey,
	object client.Object,
	opts ...client.GetOption,
) error {
	metadata, ok := object.(*metav1.PartialObjectMetadata)
	if !ok {
		return r.Reader.Get(ctx, key, object, opts...)
	}

	ownerKey := ownershipValidationOwnerKey{key: key, gvk: metadata.GroupVersionKind()}
	if cached, found := r.owners[ownerKey]; found {
		cached.DeepCopyInto(metadata)
		return nil
	}
	if err := r.Reader.Get(ctx, key, metadata, opts...); err != nil {
		return err
	}
	r.owners[ownerKey] = metadata.DeepCopy()
	return nil
}

func validateObjectOwner(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	object client.Object,
	description string,
) error {
	err := topowner.VerifyStrictOwnerChain(
		ctx,
		object,
		topowner.StrictOwner{
			GroupKind: schema.GroupKind{
				Group: dromosv1alpha1.GroupVersion.Group,
				Kind:  dromosv1alpha1.WorkloadRunnerKind,
			},
			Namespace: runner.Namespace,
			Name:      runner.Name,
			UID:       runner.UID,
		},
		reader,
	)
	if err == nil {
		return nil
	}
	reason := fmt.Sprintf(
		"%s %s/%s is not owned by WorkloadRunner %s/%s",
		description, object.GetNamespace(), object.GetName(), runner.Namespace, runner.Name,
	)
	if apierrors.IsNotFound(err) || apierrors.IsForbidden(err) || apimeta.IsNoMatchError(err) ||
		errors.Is(err, topowner.ErrInvalidOwnerChain) {
		return ownershipViolationWithCausef(err, "%s", reason)
	}
	return fmt.Errorf("%s: %w", reason, err)
}

func ownershipViolationf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errOwnershipViolation, fmt.Sprintf(format, args...))
}

func ownershipViolationWithCausef(cause error, format string, args ...any) error {
	return fmt.Errorf("%w: %s: %w", errOwnershipViolation, fmt.Sprintf(format, args...), cause)
}

// prioritizeOwnershipErrors returns only the highest-severity category so that
// callers cannot mistake a denied graph for a transient or not-yet-created one.
// Within a category, all discovered errors are retained for diagnostics.
func prioritizeOwnershipErrors(errs ...error) error {
	violations := make([]error, 0)
	transient := make([]error, 0)
	notFound := make([]error, 0)
	for _, err := range errs {
		if err == nil {
			continue
		}
		switch {
		case errors.Is(err, errOwnershipViolation):
			violations = append(violations, err)
		case apierrors.IsNotFound(err):
			notFound = append(notFound, err)
		default:
			transient = append(transient, err)
		}
	}
	if len(violations) > 0 {
		return joinOwnershipErrors(violations)
	}
	if len(transient) > 0 {
		return joinOwnershipErrors(transient)
	}
	return joinOwnershipErrors(notFound)
}

func joinOwnershipErrors(errs []error) error {
	if len(errs) == 1 {
		return errs[0]
	}
	return errors.Join(errs...)
}
