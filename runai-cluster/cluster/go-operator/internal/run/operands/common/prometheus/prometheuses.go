package prometheus

import (
	kai_consts "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// Configures shared Prometheus settings across different cluster types: standalone clusters, management clusters, and tenant clusters.
// `prom` and `runaiConfig.Spec.Prometheus` must be set.
func SetCommonPrometheusFields(prom *monitoringv1.Prometheus, runaiConfig *runaiv1.RunaiConfig, monitoredNamespaces []string, handleExternalLabels bool, enableOTLPReceiver bool) {
	prom.Spec = monitoringv1.PrometheusSpec(*runaiConfig.Spec.Prometheus.Spec.DeepCopy())

	// Configure service monitor selectors
	prom.Spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "kubernetes.io/metadata.name",
				Operator: metav1.LabelSelectorOpIn,
				Values:   monitoredNamespaces,
			},
		},
	}

	prom.Spec.ServiceMonitorSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      kai_consts.DefaultAccountingLabelKey,
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{kai_consts.DefaultAccountingLabelValue},
			},
		},
	}

	// Configure pod monitor selectors
	prom.Spec.PodMonitorNamespaceSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "kubernetes.io/metadata.name",
				Operator: metav1.LabelSelectorOpIn,
				Values:   monitoredNamespaces,
			},
		},
	}
	prom.Spec.PodMonitorSelector = &metav1.LabelSelector{}

	// Configure RuleSelector and RuleNamespaceSelector
	prom.Spec.RuleNamespaceSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "kubernetes.io/metadata.name",
				Operator: metav1.LabelSelectorOpIn,
				Values:   monitoredNamespaces,
			},
		},
	}
	prom.Spec.RuleSelector = &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      kai_consts.DefaultAccountingLabelKey,
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{kai_consts.DefaultAccountingLabelValue},
			},
		},
	}

	// Configure Affinity and Tolerations
	if promAffinity := runaiConfig.Spec.Prometheus.Spec.Affinity; promAffinity == nil {
		prom.Spec.Affinity = utils.GetAffinity(runaiConfig)
		prom.Spec.Tolerations = utils.GetTolerations(runaiConfig)
	}

	prom.Spec.Replicas = runaiConfig.Spec.Prometheus.Spec.Replicas

	// Configure SecurityContext
	prom.Spec.SecurityContext = &v1.PodSecurityContext{
		RunAsNonRoot: ptr.To(true),
		SeccompProfile: &v1.SeccompProfile{
			Type: v1.SeccompProfileTypeRuntimeDefault,
		},
		RunAsUser: ptr.To[int64](10000),
	}

	// Configure ExternalLabels
	if handleExternalLabels {
		prom.Spec.ExternalLabels = utils.InitMapIfNil(prom.Spec.ExternalLabels)
		_, exists := prom.Spec.ExternalLabels["prometheus"]
		if !exists {
			prom.Spec.ExternalLabels["prometheus"] = ""
		}
		_, exists = prom.Spec.ExternalLabels["prometheus_replica"]
		if !exists {
			prom.Spec.ExternalLabels["prometheus_replica"] = ""
		}
		_, exists = prom.Spec.ExternalLabels["__replica__"]
		if !exists {
			prom.Spec.ExternalLabels["__replica__"] = "$(POD_NAME)"
		}
		_, exists = prom.Spec.ExternalLabels["cluster"]
		if !exists {
			prom.Spec.ExternalLabels["cluster"] = runaiConfig.Spec.Internal.Global.ControlPlane.ClusterUUID
		}

		if prom.Spec.ReplicaExternalLabelName == nil {
			prom.Spec.ReplicaExternalLabelName = ptr.To("__replica__")
		}
	}

	prom.Spec.ImagePullSecrets = utils.GetGlobalImagePullSecrets(runaiConfig)

	if enableOTLPReceiver {
		prom.Spec.EnableOTLPReceiver = ptr.To(true)
	}
}
