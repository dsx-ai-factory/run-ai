package pre_hooks

import (
	"context"
	"encoding/json"
	"fmt"

	cpclusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func addPrivilegedPodSecurityStandardToRunaiNamespace(k8s client.Client) error {
	return labelNamespace(k8s, runaiv1.DefaultNamespace, "pod-security.kubernetes.io/enforce", "privileged")
}

func labelNamespace(k8s client.Client, namespace, label, value string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	patchData, err := json.Marshal(&struct {
		Metadata metav1.ObjectMeta `json:"metadata"`
	}{
		Metadata: metav1.ObjectMeta{
			Labels: map[string]string{
				label: value,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal patch data: %w", err)
	}

	err = k8s.Patch(context.Background(), ns, client.RawPatch(types.MergePatchType, patchData))
	if err != nil {
		return fmt.Errorf("failed to patch runai namespace: %w", err)
	}
	return nil
}

func (p *PreUpgradeHook) dynamicRunaiConfigUpgrade(ctx context.Context, clients *clients.Clients) error {
	resourceGroupsUpgrader := NewResourceGroupUpgrader()
	return resourceGroupsUpgrader.Upgrade(ctx, clients)
}

func (p *PreUpgradeHook) verifyRunaiPublicConfigMapExist(ctx context.Context, k8s dynamic.Interface) error {
	var cmResource = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}
	_, err := k8s.Resource(cmResource).Namespace(runaiv1.DefaultNamespace).Get(ctx, runaiv1.RunaiPublicConfigMapName, metav1.GetOptions{})
	return err
}

// Upgrade is not allowed for runai version 2.8 or lower.
// However, we don't store the runai version on the operator.
// runai-public config map was added only after 2.8, so we use it as an indication.
func (p *PreUpgradeHook) isUpgradeAllowed(ctx context.Context, k8s dynamic.Interface) error {
	runaiConfigFound, runaiConfigFoundErr := ra.RunaiConfigExists(ctx, k8s)
	runaiPublicConfigMapFoundErr := p.verifyRunaiPublicConfigMapExist(ctx, k8s)

	if runaiConfigFoundErr == nil && runaiConfigFound && runaiPublicConfigMapFoundErr == nil {
		// if both are found - upgrade is allowed. Probably installation failure...
		return nil
	}

	runaiConfigNotFound := runaiConfigFoundErr == nil && !runaiConfigFound
	runaiPublicConfigMapNotFound := runaiPublicConfigMapFoundErr != nil && k8sErrors.IsNotFound(runaiPublicConfigMapFoundErr)

	if runaiConfigNotFound && runaiPublicConfigMapNotFound {
		// if both not found - upgrade is allowed.
		// It may happen in a case of a failed uninstallation attempt, a fresh installation,
		// or if someone manually deleted runaiconfig
		return nil
	}

	if runaiConfigFoundErr != nil {
		return runaiConfigFoundErr
	}
	return runaiPublicConfigMapFoundErr
}

func headlessClusterPlaceholderControlPlaneConfig() *internal.Internal {
	return &internal.Internal{
		Global: &internal.Global{
			ClusterURL: "headless",
			Platform: &internal.Platform{
				KubernetesFlavor: internal.KubernetesFlavorVanilla,
				Version:          "v1.32.0",
			},
			ControlPlane: &internal.ControlPlane{
				Config: &internal.ClusterControlPlaneConfig{
					ClusterControlPlaneConfig: cpclusterclient.ClusterControlPlaneConfig{
						Agent: &cpclusterclient.ClusterControlPlaneConfigAgent{
							Client: &cpclusterclient.ClusterControlPlaneConfigAgentClient{
								Id:     ptr.To("headless"),
								Name:   ptr.To("headless"),
								Secret: ptr.To("headless"),
							},
						},
						Workloads: &cpclusterclient.ClusterControlPlaneConfigWorkloads{
							Client: &cpclusterclient.ClusterControlPlaneConfigWorkloadsClient{
								Id:     ptr.To("headless"),
								Name:   ptr.To("headless"),
								Secret: ptr.To("headless"),
							},
						},
						MetricsRemoteWrite: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
							Label: ptr.To("headless"),
							Url:   ptr.To("headless"),
							Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
								Username: ptr.To("headless"),
								Password: ptr.To("headless"),
							},
						},
						MetricsRemoteWrites: []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
							{
								Label: ptr.To("headless"),
								Url:   ptr.To("headless"),
								Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
									Username: ptr.To("headless"),
									Password: ptr.To("headless"),
								},
							},
						},
						ResearcherService: &cpclusterclient.ClusterControlPlaneConfigResearcherService{
							JwksURL:    ptr.To("headless"),
							LogoutLink: ptr.To("headless"),
						},
						ClusterName: ptr.To("headless"),
					},
				},
				URL:         "headless",
				ClusterUUID: "headless",
			},
			Version: "",
		},
		WorkloadController: &internal.WorkloadController{},
		ResearcherService: &internal.ResearcherService{
			Ingress: &internal.ResearcherServiceIngress{
				Enabled: ptr.To(false),
			},
		},
		Development: &internal.Development{
			Headless:             true,
			SkipDependencyChecks: true,
		},
	}
}
