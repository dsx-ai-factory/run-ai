package orchestrator2

import (
	"context"
	"strings"
	"time"

	authzv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// gatewayGVR is the Gateway API Gateway resource used to read a Gateway's
// readiness for the external-connection preflight check.
var gatewayGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "gateways",
}

const gatewayAPIGroup = "gateway.networking.k8s.io"

// GatewayStatus is the readiness of a Gateway API Gateway, distilled from its
// status conditions. Found is false when the Gateway does not exist.
type GatewayStatus struct {
	Found      bool
	Accepted   bool
	Programmed bool
}

const nodeLeaseNamespace = "kube-node-lease"

// ResourceAttributes identifies a single Kubernetes action for a permission
// check, mirroring authorization/v1 ResourceAttributes.
type ResourceAttributes struct {
	Verb      string
	Group     string
	Resource  string
	Namespace string
}

// NodeInfo is the minimal, typed view of a cluster node that preflight checks
// need: its name, its roles (derived from node-role.kubernetes.io/* labels) and
// its total (capacity) CPU/memory/ephemeral-storage.
type NodeInfo struct {
	Name   string
	Roles  []string
	CPU    resource.Quantity
	Memory resource.Quantity
	Disk   resource.Quantity
	// Labels are the node's labels, used by the system-node selection to match a
	// label selector.
	Labels map[string]string
	// Unschedulable reflects the node's .spec.unschedulable (cordoned), so the
	// system-node check can ignore nodes that cannot host pods.
	Unschedulable bool
}

// NodeHeartbeat is a node's last lease renewal: the kubelet stamps RenewTime
// with the node's own clock, so comparing renew times across nodes reveals
// clock skew (and, for very stale ones, an unreachable/dead node).
type NodeHeartbeat struct {
	Name      string
	RenewTime time.Time
}

// ClusterProbe is the read-only cluster capability preflight checks depend on.
// Everything it exposes is a non-mutating API call (discovery, access review),
// so it is safe to run in any mode, including dry-run.
type ClusterProbe interface {
	// ServerVersion returns the cluster's reported Kubernetes version and
	// doubles as a reachability/authn probe.
	ServerVersion(ctx context.Context) (string, error)
	// CanI reports whether the caller may perform the given action, via a
	// SelfSubjectAccessReview (no writes).
	CanI(ctx context.Context, ra ResourceAttributes) (allowed bool, reason string, err error)
	// ListNodes returns the cluster's nodes with their roles and total
	// (capacity) resources, for node-sizing preflight checks.
	ListNodes(ctx context.Context) ([]NodeInfo, error)
	// NodeHeartbeats returns each node's last lease renewal (kube-node-lease),
	// for the clock-skew preflight check.
	NodeHeartbeats(ctx context.Context) ([]NodeHeartbeat, error)
	// HasDefaultStorageClass reports whether the cluster has a StorageClass
	// marked default, so PVCs without an explicit storageClassName can be
	// dynamically provisioned. Used by the storage preflight check.
	HasDefaultStorageClass(ctx context.Context) (bool, error)
	// NamespaceExists reports whether a namespace is present, so a check can
	// confirm a stack that lives in a well-known namespace is installed (e.g.
	// the OpenShift monitoring stack in openshift-monitoring).
	NamespaceExists(ctx context.Context, name string) (bool, error)
	// IngressClassExists reports whether an IngressClass of the given name is
	// present, so the external-connection check can confirm the configured
	// ingress controller (nginx, haproxy, ...) is installed.
	IngressClassExists(ctx context.Context, name string) (bool, error)
	// GatewayAPIInstalled reports whether the Gateway API is installed on the
	// cluster (the gateway.networking.k8s.io/v1 API is served).
	GatewayAPIInstalled(ctx context.Context) (bool, error)
	// GatewayStatus returns the readiness of a Gateway API Gateway (whether it
	// exists and its Accepted/Programmed conditions).
	GatewayStatus(ctx context.Context, namespace, name string) (GatewayStatus, error)
}

// defaultStorageClassAnnotations are the (stable and beta) annotations that mark
// a StorageClass as the cluster default.
var defaultStorageClassAnnotations = []string{
	"storageclass.kubernetes.io/is-default-class",
	"storageclass.beta.kubernetes.io/is-default-class",
}

type realClusterProbe struct {
	disco discovery.DiscoveryInterface
	authz kubernetes.Interface
	dyn   dynamic.Interface
}

// NewClusterProbe builds the production ClusterProbe from a REST config.
func NewClusterProbe(cfg *rest.Config) (ClusterProbe, error) {
	disco, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &realClusterProbe{disco: disco, authz: cs, dyn: dyn}, nil
}

func (p *realClusterProbe) ServerVersion(_ context.Context) (string, error) {
	v, err := p.disco.ServerVersion()
	if err != nil {
		return "", err
	}
	return v.GitVersion, nil
}

func (p *realClusterProbe) CanI(ctx context.Context, ra ResourceAttributes) (bool, string, error) {
	review := &authzv1.SelfSubjectAccessReview{
		Spec: authzv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authzv1.ResourceAttributes{
				Verb:      ra.Verb,
				Group:     ra.Group,
				Resource:  ra.Resource,
				Namespace: ra.Namespace,
			},
		},
	}
	resp, err := p.authz.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, "", err
	}
	return resp.Status.Allowed, resp.Status.Reason, nil
}

func (p *realClusterProbe) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	list, err := p.authz.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodes := make([]NodeInfo, 0, len(list.Items))
	for _, n := range list.Items {
		nodes = append(nodes, NodeInfo{
			Name:          n.Name,
			Roles:         nodeRoles(n.Labels),
			CPU:           n.Status.Capacity[corev1.ResourceCPU],
			Memory:        n.Status.Capacity[corev1.ResourceMemory],
			Disk:          n.Status.Capacity[corev1.ResourceEphemeralStorage],
			Labels:        n.Labels,
			Unschedulable: n.Spec.Unschedulable,
		})
	}
	return nodes, nil
}

func (p *realClusterProbe) NodeHeartbeats(ctx context.Context) ([]NodeHeartbeat, error) {
	leases, err := p.authz.CoordinationV1().Leases(nodeLeaseNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	beats := make([]NodeHeartbeat, 0, len(leases.Items))
	for _, l := range leases.Items {
		if l.Spec.RenewTime == nil {
			continue
		}
		beats = append(beats, NodeHeartbeat{Name: l.Name, RenewTime: l.Spec.RenewTime.Time})
	}
	return beats, nil
}

func (p *realClusterProbe) HasDefaultStorageClass(ctx context.Context) (bool, error) {
	list, err := p.authz.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, sc := range list.Items {
		for _, ann := range defaultStorageClassAnnotations {
			if sc.Annotations[ann] == "true" {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *realClusterProbe) NamespaceExists(ctx context.Context, name string) (bool, error) {
	_, err := p.authz.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *realClusterProbe) IngressClassExists(ctx context.Context, name string) (bool, error) {
	_, err := p.authz.NetworkingV1().IngressClasses().Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *realClusterProbe) GatewayAPIInstalled(_ context.Context) (bool, error) {
	groups, err := p.disco.ServerGroups()
	if err != nil {
		return false, err
	}
	for _, g := range groups.Groups {
		if g.Name != gatewayAPIGroup {
			continue
		}
		for _, v := range g.Versions {
			if v.Version == gatewayGVR.Version {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *realClusterProbe) GatewayStatus(ctx context.Context, namespace, name string) (GatewayStatus, error) {
	gw, err := p.dyn.Resource(gatewayGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return GatewayStatus{Found: false}, nil
	}
	if err != nil {
		return GatewayStatus{}, err
	}
	status := GatewayStatus{Found: true}
	conditions, _, _ := unstructured.NestedSlice(gw.Object, "status", "conditions")
	for _, c := range conditions {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		condType, _ := cond["type"].(string)
		condStatus, _ := cond["status"].(string)
		switch condType {
		case "Accepted":
			status.Accepted = condStatus == "True"
		case "Programmed":
			status.Programmed = condStatus == "True"
		}
	}
	return status, nil
}

// nodeRoles extracts role names from node-role.kubernetes.io/<role> labels.
func nodeRoles(labels map[string]string) []string {
	const prefix = "node-role.kubernetes.io/"
	var roles []string
	for k := range labels {
		if role := strings.TrimPrefix(k, prefix); role != k && role != "" {
			roles = append(roles, role)
		}
	}
	return roles
}
