package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/shared_data"
	prometheus "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/prometheus/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	v1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	prometheusv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Prometheus Operand", func() {
	BeforeEach(func() {
		runaiConfig.Spec.Prometheus.Enabled = ptr.To[bool](true)
		EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
		PollForTrueCondition(k8s, v1.ConditionTypeDeployed, runaiConfig)
	})

	AfterEach(func() {
		runaiConfig.Spec.Prometheus.Enabled = ptr.To[bool](true)
		EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
		PollForTrueCondition(k8s, v1.ConditionTypeDeployed, runaiConfig)
	})

	Context("OpenShift configuration", func() {
		BeforeEach(func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: consts.OpenshiftMonitoringNamespace,
				},
			}
			ExpectCreateResource(k8s, ns)

			runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runv1_internal.KubernetesFlavorOpenShift
			EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
			PollForTrueCondition(k8s, v1.ConditionTypeDeployed, runaiConfig)
		})

		AfterEach(func() {
			runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runv1_internal.KubernetesFlavorOpenShift
			EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
			PollForTrueCondition(k8s, v1.ConditionTypeDeployed, runaiConfig)

			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: consts.OpenshiftMonitoringNamespace,
				},
			}
			DeleteAndPollUntilDeleted(k8s, ns)
		})

		It("should update openshift cluster-monitoring configmap", func() {
			res, err := resources.NewConfigMap("prometheus", kc.New(k8s, k8s, nil), runaiConfig).Object(apiCtx)
			ExpectNoErr(err)

			cm := res.(*corev1.ConfigMap)

			Expect(cm.Data["config.yaml"]).To(ContainSubstring("externalLabels"))
		})
	})

	Context("Resources", func() {
		It("should have resources", func() {
			operand := prometheus.Operand(kc.New(k8s, k8s, nil), runaiConfig, shared_data.NewSharedData())
			resources := operand.Resources()
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			operand := prometheus.Operand(kc.New(k8s, k8s, nil), runaiConfig, shared_data.NewSharedData())
			resources := operand.Resources()

			for _, resource := range resources {
				res, err := resource.Object(apiCtx)
				ExpectNoErr(err)
				ExpectGetResource(k8s, res)
			}
		})

		It("should re-create deleted resources", func() {
			operand := prometheus.Operand(kc.New(k8s, k8s, nil), runaiConfig, shared_data.NewSharedData())
			resources := operand.Resources()

			for _, resource := range resources {
				res, err := resource.Object(apiCtx)
				ExpectNoErr(err)
				DeleteAndPollUntilDeleted(k8s, res)
			}

			EventuallyWithTimeout(func() bool {
				reCreated := true
				for _, resource := range resources {
					res, err := resource.Object(apiCtx)
					ExpectNoErr(err)
					if !ResourceExists(k8s, res) {
						reCreated = false
					}
				}

				return reCreated
			}).Should(BeTrue())
		})
	})

	Context("Modified resources", func() {
		It("should reconcile a modified Prometheus resource", func() {
			runaiConfig.Spec.Prometheus.Spec.RemoteWrite = []monitoringv1.RemoteWriteSpec{
				{
					RemoteTimeout: ptr.To(monitoringv1.Duration("10s")),
				},
			}
			EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
			WaitForOperatorToObserve(kc.New(k8s, k8s, nil), runaiConfig)

			res, err := resources.NewPrometheusResource("prometheus", kc.New(k8s, k8s, nil), runaiConfig).Object(apiCtx)
			ExpectNoErr(err)

			prom := res.(*monitoringv1.Prometheus)
			prom.Spec.RemoteWrite[0].RemoteTimeout = ptr.To(monitoringv1.Duration("0s"))
			EventuallyUpdateResource(k8s, prom)

			EventuallyWithTimeout(func() bool {
				ExpectGetResource(k8s, prom)
				return prom.Spec.RemoteWrite[0].RemoteTimeout == ptr.To(monitoringv1.Duration("10s"))
			}).Should(BeTrue())
		})

		It("should reconcile a modified PrometheusRule resource", func() {
			runaiConfig.Spec.Prometheus.RuleGroups = prometheusv1.RuleGroups{
				{
					Name:  "real-name",
					Rules: []monitoringv1.Rule{},
				},
			}
			EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
			WaitForDeployedCondition(kc.New(k8s, k8s, nil), runaiConfig)

			res, err := resources.NewPrometheusRuleResource("prometheus", kc.New(k8s, k8s, nil), runaiConfig).Object(apiCtx)
			ExpectNoErr(err)

			originalRule := res.(*monitoringv1.PrometheusRule)
			rule := originalRule.DeepCopy()

			rule.Spec.Groups[0].Name = "test"
			EventuallyUpdateResource(k8s, rule)
			WaitForDeployedCondition(k8s, GetTestingRunaiConfig(k8s))

			EventuallyWithTimeout(func() bool {
				ExpectGetResource(k8s, rule)
				rule := res.(*monitoringv1.PrometheusRule)
				return rule.Spec.Groups[0].Name == "real-name"
			}).Should(BeTrue())
		})
	})
})
