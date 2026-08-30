package dra

import (
	"context"

	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	resourcev1client "k8s.io/client-go/kubernetes/typed/resource/v1"
	resourcev1beta1client "k8s.io/client-go/kubernetes/typed/resource/v1beta1"
	resourcev1beta2client "k8s.io/client-go/kubernetes/typed/resource/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// cachedKubeClient wraps kubernetes.Interface to serve ResourceClaim and
// ResourceClaimTemplate Get calls from a cached client.Reader.
// Supports v1, v1beta1, and v1beta2 API versions.
//
// IMPORTANT: All Get calls MUST go through the cachedReader, never through the
// embedded kubernetes.Interface. The leaf wrapper types intentionally embed nil
// interfaces so that any non-overridden method panics immediately instead of
// silently using the uncached client.
type cachedKubeClient struct {
	kubernetes.Interface
	cachedReader client.Reader
}

func newCachedKubeClient(cachedReader client.Reader, kubeClient kubernetes.Interface) *cachedKubeClient {
	return &cachedKubeClient{
		Interface:    kubeClient,
		cachedReader: cachedReader,
	}
}

// convertNoMatchToNotFound converts meta.NoKindMatchError to apierrors.NotFound.
// draclient's version negotiation uses apierrors.IsNotFound to decide whether to
// try the next API version (v1 → v1beta2 → v1beta1). The cached reader returns
// meta.NoKindMatchError when the GVK is not in the cache (e.g. cluster supports
// v1beta2 but not v1), which draclient does not recognize. Converting it to
// NotFound allows the version fallback to work correctly.
func convertNoMatchToNotFound(err error, resource string, name string) error {
	if meta.IsNoMatchError(err) {
		return apierrors.NewNotFound(schema.GroupResource{Resource: resource}, name)
	}
	return err
}

func (c *cachedKubeClient) ResourceV1() resourcev1client.ResourceV1Interface {
	return &cachedResourceV1{
		ResourceV1Interface: c.Interface.ResourceV1(),
		cachedReader:        c.cachedReader,
	}
}

func (c *cachedKubeClient) ResourceV1beta1() resourcev1beta1client.ResourceV1beta1Interface {
	return &cachedResourceV1beta1{
		ResourceV1beta1Interface: c.Interface.ResourceV1beta1(),
		cachedReader:             c.cachedReader,
	}
}

func (c *cachedKubeClient) ResourceV1beta2() resourcev1beta2client.ResourceV1beta2Interface {
	return &cachedResourceV1beta2{
		ResourceV1beta2Interface: c.Interface.ResourceV1beta2(),
		cachedReader:             c.cachedReader,
	}
}

// --- v1 wrappers ---

type cachedResourceV1 struct {
	resourcev1client.ResourceV1Interface
	cachedReader client.Reader
}

func (c *cachedResourceV1) ResourceClaims(namespace string) resourcev1client.ResourceClaimInterface {
	return &cachedResourceClaimsV1{
		// ResourceClaimInterface intentionally nil — panics on any non-overridden method
		cachedReader: c.cachedReader,
		namespace:    namespace,
	}
}

func (c *cachedResourceV1) ResourceClaimTemplates(namespace string) resourcev1client.ResourceClaimTemplateInterface {
	return &cachedResourceClaimTemplatesV1{
		// ResourceClaimTemplateInterface intentionally nil — panics on any non-overridden method
		cachedReader: c.cachedReader,
		namespace:    namespace,
	}
}

type cachedResourceClaimsV1 struct {
	resourcev1client.ResourceClaimInterface
	cachedReader client.Reader
	namespace    string
}

func (c *cachedResourceClaimsV1) Get(ctx context.Context, name string, opts metav1.GetOptions) (*resourcev1.ResourceClaim, error) {
	claim := &resourcev1.ResourceClaim{}
	if err := c.cachedReader.Get(ctx, client.ObjectKey{Namespace: c.namespace, Name: name}, claim); err != nil {
		return nil, convertNoMatchToNotFound(err, "resourceclaims", name)
	}
	return claim, nil
}

type cachedResourceClaimTemplatesV1 struct {
	resourcev1client.ResourceClaimTemplateInterface
	cachedReader client.Reader
	namespace    string
}

func (c *cachedResourceClaimTemplatesV1) Get(ctx context.Context, name string, opts metav1.GetOptions) (*resourcev1.ResourceClaimTemplate, error) {
	template := &resourcev1.ResourceClaimTemplate{}
	if err := c.cachedReader.Get(ctx, client.ObjectKey{Namespace: c.namespace, Name: name}, template); err != nil {
		return nil, convertNoMatchToNotFound(err, "resourceclaimtemplates", name)
	}
	return template, nil
}

// --- v1beta1 wrappers ---

type cachedResourceV1beta1 struct {
	resourcev1beta1client.ResourceV1beta1Interface
	cachedReader client.Reader
}

func (c *cachedResourceV1beta1) ResourceClaims(namespace string) resourcev1beta1client.ResourceClaimInterface {
	return &cachedResourceClaimsV1beta1{
		// ResourceClaimInterface intentionally nil — panics on any non-overridden method
		cachedReader: c.cachedReader,
		namespace:    namespace,
	}
}

func (c *cachedResourceV1beta1) ResourceClaimTemplates(namespace string) resourcev1beta1client.ResourceClaimTemplateInterface {
	return &cachedResourceClaimTemplatesV1beta1{
		// ResourceClaimTemplateInterface intentionally nil — panics on any non-overridden method
		cachedReader: c.cachedReader,
		namespace:    namespace,
	}
}

type cachedResourceClaimsV1beta1 struct {
	resourcev1beta1client.ResourceClaimInterface
	cachedReader client.Reader
	namespace    string
}

func (c *cachedResourceClaimsV1beta1) Get(ctx context.Context, name string, opts metav1.GetOptions) (*resourcev1beta1.ResourceClaim, error) {
	claim := &resourcev1beta1.ResourceClaim{}
	if err := c.cachedReader.Get(ctx, client.ObjectKey{Namespace: c.namespace, Name: name}, claim); err != nil {
		return nil, convertNoMatchToNotFound(err, "resourceclaims", name)
	}
	return claim, nil
}

type cachedResourceClaimTemplatesV1beta1 struct {
	resourcev1beta1client.ResourceClaimTemplateInterface
	cachedReader client.Reader
	namespace    string
}

func (c *cachedResourceClaimTemplatesV1beta1) Get(ctx context.Context, name string, opts metav1.GetOptions) (*resourcev1beta1.ResourceClaimTemplate, error) {
	template := &resourcev1beta1.ResourceClaimTemplate{}
	if err := c.cachedReader.Get(ctx, client.ObjectKey{Namespace: c.namespace, Name: name}, template); err != nil {
		return nil, convertNoMatchToNotFound(err, "resourceclaimtemplates", name)
	}
	return template, nil
}

// --- v1beta2 wrappers ---

type cachedResourceV1beta2 struct {
	resourcev1beta2client.ResourceV1beta2Interface
	cachedReader client.Reader
}

func (c *cachedResourceV1beta2) ResourceClaims(namespace string) resourcev1beta2client.ResourceClaimInterface {
	return &cachedResourceClaimsV1beta2{
		// ResourceClaimInterface intentionally nil — panics on any non-overridden method
		cachedReader: c.cachedReader,
		namespace:    namespace,
	}
}

func (c *cachedResourceV1beta2) ResourceClaimTemplates(namespace string) resourcev1beta2client.ResourceClaimTemplateInterface {
	return &cachedResourceClaimTemplatesV1beta2{
		// ResourceClaimTemplateInterface intentionally nil — panics on any non-overridden method
		cachedReader: c.cachedReader,
		namespace:    namespace,
	}
}

type cachedResourceClaimsV1beta2 struct {
	resourcev1beta2client.ResourceClaimInterface
	cachedReader client.Reader
	namespace    string
}

func (c *cachedResourceClaimsV1beta2) Get(ctx context.Context, name string, opts metav1.GetOptions) (*resourcev1beta2.ResourceClaim, error) {
	claim := &resourcev1beta2.ResourceClaim{}
	if err := c.cachedReader.Get(ctx, client.ObjectKey{Namespace: c.namespace, Name: name}, claim); err != nil {
		return nil, convertNoMatchToNotFound(err, "resourceclaims", name)
	}
	return claim, nil
}

type cachedResourceClaimTemplatesV1beta2 struct {
	resourcev1beta2client.ResourceClaimTemplateInterface
	cachedReader client.Reader
	namespace    string
}

func (c *cachedResourceClaimTemplatesV1beta2) Get(ctx context.Context, name string, opts metav1.GetOptions) (*resourcev1beta2.ResourceClaimTemplate, error) {
	template := &resourcev1beta2.ResourceClaimTemplate{}
	if err := c.cachedReader.Get(ctx, client.ObjectKey{Namespace: c.namespace, Name: name}, template); err != nil {
		return nil, convertNoMatchToNotFound(err, "resourceclaimtemplates", name)
	}
	return template, nil
}
