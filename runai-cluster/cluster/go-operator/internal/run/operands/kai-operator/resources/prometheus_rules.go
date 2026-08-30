package resources

import (
	"context"
	"fmt"

	kai_consts "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const prometheusRuleName = "runai-accounting"

func PrometheusRuleForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	rule := &monitoringv1.PrometheusRule{}
	err := k8sReader.Get(ctx, prometheusRuleKey(runaiConfig), rule)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	desired := desiredPrometheusRule(runaiConfig)
	rule.Name = desired.Name
	rule.Namespace = desired.Namespace
	rule.Labels = desired.Labels
	rule.Spec = desired.Spec
	return rule, nil
}

func prometheusRuleKey(runaiConfig *runv1.RunaiConfig) types.NamespacedName {
	return types.NamespacedName{Namespace: runaiConfig.Namespace, Name: prometheusRuleName}
}

func desiredPrometheusRule(runaiConfig *runv1.RunaiConfig) *monitoringv1.PrometheusRule {
	rule := &monitoringv1.PrometheusRule{}
	rule.Name = prometheusRuleName
	rule.Namespace = runaiConfig.Namespace
	rule.Labels = map[string]string{
		kai_consts.DefaultAccountingLabelKey: kai_consts.DefaultAccountingLabelValue,
	}
	rule.Spec.Groups = []monitoringv1.RuleGroup{{
		Name: "default",
		Rules: []monitoringv1.Rule{{
			Record: "resources_per_nodepool",
			Expr:   intstr.FromString("sum(kube_node_status_capacity * on(node) group_left(nodepool) (max by (node, nodepool) (runai_node_nodepool == 1))) by (nodepool, resource)"),
		}},
	}}
	return rule
}

func PrometheusRuleStatus(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	rule := &monitoringv1.PrometheusRule{}
	err := k8sReader.Get(ctx, prometheusRuleKey(runaiConfig), rule)
	if errors.IsNotFound(err) {
		return operands.NotReadyStatus(fmt.Sprintf("prometheus rule [%s] is missing", prometheusRuleName)), nil
	}
	if err != nil {
		return operands.NotReadyStatus(), err
	}
	if rule.UID == "" {
		return operands.NotReadyStatus(fmt.Sprintf("prometheus rule [%s] is missing", rule.GetName())), nil
	}

	desired := desiredPrometheusRule(runaiConfig)
	if rule.Labels[kai_consts.DefaultAccountingLabelKey] != kai_consts.DefaultAccountingLabelValue ||
		!apiequality.Semantic.DeepEqual(rule.Spec, desired.Spec) {
		return operands.NotReadyStatus(fmt.Sprintf("prometheus rule [%s] is out of date", rule.GetName())), nil
	}
	return operands.ReadyStatus(), nil
}
