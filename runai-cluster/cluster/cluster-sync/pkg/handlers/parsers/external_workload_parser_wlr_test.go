/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	constants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	issuepkg "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	common "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func newWlrWithIssues(issues []dromosv1alpha1.Issue, plugins map[string]*dromosv1alpha1.PluginStatus) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		Status: dromosv1alpha1.WorkloadRunnerStatus{
			Issues:  issues,
			Plugins: plugins,
		},
	}
}

var _ = Describe("appendWlrIssues", func() {
	It("adds nothing when no issues exist", func() {
		issues := &common.Issues{}
		appendWlrIssues(issues, newWlrWithIssues(nil, nil))
		Expect(issues).To(Equal(&common.Issues{}))
	})

	It("merges runner and plugin issues, skipping nil plugins", func() {
		wlr := newWlrWithIssues(
			[]dromosv1alpha1.Issue{
				{Kind: "WorkloadRunner", Name: "wlr-1", Description: "runner error", Group: issuepkg.Internal.String()},
			},
			map[string]*dromosv1alpha1.PluginStatus{
				"networking": {
					Issues: []dromosv1alpha1.Issue{
						{Kind: "Service", Name: "svc-1", Description: "svc error", Group: issuepkg.Network.String()},
					},
				},
				"empty-plugin": {},
				"nil-plugin":   nil,
			},
		)

		issues := &common.Issues{}
		appendWlrIssues(issues, wlr)

		Expect(issues.Internal).To(HaveLen(1))
		Expect(issues.Internal[0].Name).To(Equal("wlr-1"))
		Expect(issues.Network).To(HaveLen(1))
		Expect(issues.Network[0].Name).To(Equal("svc-1"))
	})

	It("maps unknown groups to Internal by default", func() {
		wlr := newWlrWithIssues([]dromosv1alpha1.Issue{
			{Kind: "Unknown", Name: "u-1", Description: "unknown issue", Group: "UnknownGroup"},
			{Kind: "Empty", Name: "e-1", Description: "no group", Group: ""},
		}, nil)

		issues := &common.Issues{}
		appendWlrIssues(issues, wlr)

		Expect(issues.Internal).To(HaveLen(2))
	})
})

var _ = Describe("mergeWlrIssuesIntoAdditionalFields", func() {
	It("creates additional fields from nil", func() {
		issues := &common.Issues{
			Network: []common.Issue{{Kind: "Service", Name: "svc-1", Description: "error"}},
		}

		result := mergeWlrIssuesIntoAdditionalFields(nil, issues)

		Expect(result).NotTo(BeNil())
		var fields map[string]string
		Expect(json.Unmarshal([]byte(*result), &fields)).To(Succeed())
		Expect(fields).To(HaveKey("issues"))

		var parsedIssues common.Issues
		Expect(json.Unmarshal([]byte(fields["issues"]), &parsedIssues)).To(Succeed())
		Expect(parsedIssues.Network).To(HaveLen(1))
		Expect(parsedIssues.Network[0].Name).To(Equal("svc-1"))
	})

	It("preserves existing fields", func() {
		existing := `{"cliCommand":"runai submit","someKey":"someValue"}`
		issues := &common.Issues{
			Storage: []common.Issue{{Kind: "PVC", Name: "pvc-1", Description: "pending"}},
		}

		result := mergeWlrIssuesIntoAdditionalFields(&existing, issues)

		Expect(result).NotTo(BeNil())
		var fields map[string]string
		Expect(json.Unmarshal([]byte(*result), &fields)).To(Succeed())
		Expect(fields["cliCommand"]).To(Equal("runai submit"))
		Expect(fields["someKey"]).To(Equal("someValue"))
		Expect(fields).To(HaveKey("issues"))
	})

	It("overrides existing issues", func() {
		oldIssues := common.Issues{
			Internal: []common.Issue{{Kind: "Old", Name: "old-issue", Description: "old"}},
		}
		oldIssuesBytes, _ := json.Marshal(oldIssues)
		existing := `{"issues":` + string(oldIssuesBytes) + `}`

		newIssues := &common.Issues{
			Network: []common.Issue{{Kind: "New", Name: "new-issue", Description: "new"}},
		}

		result := mergeWlrIssuesIntoAdditionalFields(&existing, newIssues)

		Expect(result).NotTo(BeNil())
		var fields map[string]string
		Expect(json.Unmarshal([]byte(*result), &fields)).To(Succeed())

		var parsedIssues common.Issues
		Expect(json.Unmarshal([]byte(fields["issues"]), &parsedIssues)).To(Succeed())
		Expect(parsedIssues.Network).To(HaveLen(1))
		Expect(parsedIssues.Network[0].Name).To(Equal("new-issue"))
		Expect(parsedIssues.Internal).To(BeEmpty())
	})
})

var _ = Describe("Phase mapping", func() {
	DescribeTable("maps WLR phase to API phase",
		func(wlrPhase dromosv1alpha1.WorkloadRunnerPhase, expectedText string) {
			p := NewWorkloadParser(nil)

			result := p.getApiPhase(string(wlrPhase))

			Expect(string(result)).To(Equal(expectedText))
		},
		Entry("Running", dromosv1alpha1.WorkloadRunnerPhaseRunning, "Running"),
		Entry("Pending", dromosv1alpha1.WorkloadRunnerPhasePending, "Pending"),
		Entry("Failed", dromosv1alpha1.WorkloadRunnerPhaseFailed, "Failed"),
		Entry("Completed", dromosv1alpha1.WorkloadRunnerPhaseCompleted, "Completed"),
		Entry("Suspended -> Stopped", dromosv1alpha1.WorkloadRunnerPhaseSuspended, "Stopped"),
		Entry("Degraded", dromosv1alpha1.WorkloadRunnerPhaseDegraded, "Degraded"),
		Entry("Creating", dromosv1alpha1.WorkloadRunnerPhaseCreating, "Creating"),
		Entry("Initializing", dromosv1alpha1.WorkloadRunnerPhaseInitializing, "Initializing"),
	)

	It("maps empty phase to Creating", func() {
		p := NewWorkloadParser(nil)

		result := p.getApiPhase("")

		Expect(string(result)).To(Equal("Creating"))
	})
})

var _ = Describe("networking plugin details", func() {
	It("reads networking plugin status from additionalInfo", func() {
		networkingStatus := endpoints.NewNetworkingStatus(1)
		networkingStatus.Endpoints = []endpoints.EndpointStatus{
			{
				Services: []endpoints.ServiceStatus{
					{Name: "jupyter"},
				},
			},
		}
		raw, err := endpoints.NetworkingStatusToRawExtension(networkingStatus)
		Expect(err).To(BeNil())
		wlr := &dromosv1alpha1.WorkloadRunner{
			Status: dromosv1alpha1.WorkloadRunnerStatus{
				Plugins: map[string]*dromosv1alpha1.PluginStatus{
					networkingPluginName: {
						AdditionalInfo: raw,
					},
				},
			},
		}

		result, err := networkingDetailsFromWorkloadRunner(wlr)

		Expect(err).To(BeNil())
		Expect(result.Endpoints).To(HaveLen(1))
		Expect(result.Endpoints[0].Services).To(HaveLen(1))
		Expect(result.Endpoints[0].Services[0].Name).To(Equal("jupyter"))
	})

	It("returns no networking plugin status for a nil WorkloadRunner", func() {
		result, err := networkingDetailsFromWorkloadRunner(nil)

		Expect(err).To(BeNil())
		Expect(result).To(BeNil())
		Expect(externalUrlFromWorkloadRunner(nil)).To(BeEmpty())
		Expect(func() {
			applyNetworkingDetailsToSyncFields(endpoints.NewNetworkingStatus(1), nil)
		}).NotTo(Panic())
	})

	It("maps EWI external endpoints into endpoint sync fields", func() {
		syncEndpoints, err := parseExternalEndpoints([]v1alpha1.ExternalEndpoint{
			{
				Host:       "ewi.example.com",
				Path:       "/ready",
				Port:       "443",
				Protocol:   "https",
				Url:        "https://ewi.example.com/ready",
				TlsEnabled: ptr.To(true),
				SourceRef: v1alpha1.ExternalEndpointSourceRef{
					Kind: "Ingress",
					Name: "ewi-ingress",
				},
			},
		})

		Expect(err).To(BeNil())
		Expect(syncEndpoints).To(HaveLen(1))
		Expect(syncEndpoints[0].GetHost()).To(Equal("ewi.example.com"))
		Expect(syncEndpoints[0].GetPath()).To(Equal("/ready"))
		Expect(syncEndpoints[0].GetPort()).To(Equal(int32(443)))
		Expect(syncEndpoints[0].GetProtocol()).To(Equal("https"))
		Expect(syncEndpoints[0].GetUrl()).To(Equal("https://ewi.example.com/ready"))
		Expect(syncEndpoints[0].GetTlsEnabled()).To(BeTrue())
		Expect(syncEndpoints[0].SourceRef.GetKind()).To(Equal("Ingress"))
		Expect(syncEndpoints[0].SourceRef.GetName()).To(Equal("ewi-ingress"))
	})

	It("keeps valid EWI external endpoints when another endpoint has a malformed port", func() {
		syncEndpoints, err := parseExternalEndpoints([]v1alpha1.ExternalEndpoint{
			{
				Host:     "first.example.com",
				Port:     "443",
				Protocol: "https",
				Url:      "https://first.example.com",
				SourceRef: v1alpha1.ExternalEndpointSourceRef{
					Kind: "Ingress",
					Name: "first-ingress",
				},
			},
			{
				Host:     "bad.example.com",
				Port:     "not-a-port",
				Protocol: "https",
				Url:      "https://bad.example.com",
				SourceRef: v1alpha1.ExternalEndpointSourceRef{
					Kind: "Ingress",
					Name: "bad-ingress",
				},
			},
			{
				Host:     "second.example.com",
				Port:     "8443",
				Protocol: "https",
				Url:      "https://second.example.com",
				SourceRef: v1alpha1.ExternalEndpointSourceRef{
					Kind: "HTTPRoute",
					Name: "second-route",
				},
			},
		})

		Expect(err).To(HaveOccurred())
		Expect(syncEndpoints).To(HaveLen(2))
		Expect(syncEndpoints[0].GetHost()).To(Equal("first.example.com"))
		Expect(syncEndpoints[0].GetPort()).To(Equal(int32(443)))
		Expect(syncEndpoints[1].GetHost()).To(Equal("second.example.com"))
		Expect(syncEndpoints[1].GetPort()).To(Equal(int32(8443)))
	})

	It("maps native-parity service status and service URLs without owning endpoint sync fields", func() {
		authProxyReady := true
		ewiEndpointURL := "https://ewi.example.com/ready"
		ewiEndpoint := openapi.EndpointSync{
			Host:     ptr.To("ewi.example.com"),
			Path:     ptr.To("/ready"),
			Port:     ptr.To(int32(443)),
			Protocol: ptr.To("https"),
			Url:      ptr.To(ewiEndpointURL),
			SourceRef: &openapi.EndpointSyncAllOfSourceRef{
				Kind: ptr.To("Ingress"),
				Name: ptr.To("ewi-ingress"),
			},
		}
		ewiEndpoint.SetTlsEnabled(true)
		networkingStatus := endpoints.NewNetworkingStatus(2)
		networkingStatus.Endpoints = []endpoints.EndpointStatus{
			{
				Services: []endpoints.ServiceStatus{
					{
						Name:      "jupyter",
						Phase:     endpoints.PhaseReady,
						Reason:    "ServiceReady",
						Message:   "service jupyter exists",
						Type:      "ClusterIP",
						ClusterIP: "10.96.1.2",
						Ports: []endpoints.ServicePortStatus{
							{
								Name:       "http",
								Port:       8080,
								TargetPort: intstr.FromInt32(8888),
								Protocol:   "TCP",
							},
						},
						LoadBalancer: &endpoints.LoadBalancerStatus{
							Ingress: []endpoints.LoadBalancerIngress{
								{IP: "1.2.3.4", Hostname: "lb.example.com"},
							},
						},
						AuthProxyReady: &authProxyReady,
						URLs: []endpoints.ServiceURLStatus{
							{
								URL:        "https://jupyter.example.com/lab",
								SourceKind: endpoints.ExternalAccessKindIngress,
								SourceName: "jupyter-ingress",
							},
						},
					},
				},
			},
		}
		syncFields := &openapi.WorkloadSyncFields{
			Urls:      []string{"http://jupyter.runai.svc.cluster.local"},
			Endpoints: []openapi.EndpointSync{ewiEndpoint},
		}

		applyNetworkingDetailsToSyncFields(networkingStatus, syncFields)

		Expect(syncFields.NetworkStatus).NotTo(BeNil())
		Expect(syncFields.NetworkStatus.Services).To(HaveLen(1))
		service := syncFields.NetworkStatus.Services[0]
		Expect(service.GetName()).To(Equal("jupyter"))
		Expect(service.GetType()).To(Equal("ClusterIP"))
		Expect(service.Ports).To(HaveLen(1))
		Expect(service.Ports[0].GetName()).To(Equal("http"))
		Expect(service.Ports[0].GetPort()).To(Equal(int32(8080)))
		Expect(service.Ports[0].GetTargetPort()).To(Equal("8888"))
		Expect(service.Ports[0].GetProtocol()).To(Equal("TCP"))
		Expect(service.LoadBalancer.GetIngress()).To(HaveLen(1))
		Expect(service.LoadBalancer.GetIngress()[0].GetIp()).To(Equal("1.2.3.4"))
		Expect(service.LoadBalancer.GetIngress()[0].GetHostname()).To(Equal("lb.example.com"))
		Expect(service.Urls).To(HaveLen(1))
		Expect(service.Urls[0].GetUrl()).To(Equal("https://jupyter.example.com/lab"))
		Expect(service.Urls[0].GetSourceKind()).To(Equal("Ingress"))
		Expect(service.Urls[0].GetSourceName()).To(Equal("jupyter-ingress"))
		Expect(syncFields.Urls).To(Equal([]string{"https://jupyter.example.com/lab"}))
		Expect(syncFields.Endpoints).To(HaveLen(1))
		Expect(syncFields.Endpoints[0].GetUrl()).To(Equal(ewiEndpointURL))
		Expect(syncFields.Endpoints[0].GetHost()).To(Equal("ewi.example.com"))
		Expect(syncFields.Endpoints[0].GetPath()).To(Equal("/ready"))
		Expect(syncFields.Endpoints[0].GetPort()).To(Equal(int32(443)))
		Expect(syncFields.Endpoints[0].GetProtocol()).To(Equal("https"))
		Expect(syncFields.Endpoints[0].GetTlsEnabled()).To(BeTrue())
		Expect(syncFields.Endpoints[0].SourceRef.GetKind()).To(Equal("Ingress"))
		Expect(syncFields.Endpoints[0].SourceRef.GetName()).To(Equal("ewi-ingress"))
	})

	It("adds EWI service endpoints to matching NodePort service status URLs", func() {
		networkingStatus := endpoints.NewNetworkingStatus(1)
		networkingStatus.Endpoints = []endpoints.EndpointStatus{
			{
				Services: []endpoints.ServiceStatus{
					{
						Name: "svc-nodeport",
						Type: "NodePort",
						Ports: []endpoints.ServicePortStatus{
							{
								Name:       "jupyter-port",
								Port:       32101,
								TargetPort: intstr.FromInt32(8888),
								Protocol:   "TCP",
							},
							{
								Name:       "tensorboard-port",
								Port:       32102,
								TargetPort: intstr.FromInt32(6006),
								Protocol:   "TCP",
							},
						},
					},
					{
						Name: "svc-clusterip",
						Type: "ClusterIP",
					},
				},
			},
		}
		syncFields := &openapi.WorkloadSyncFields{
			Urls: []string{"https://custom.example.com"},
			Endpoints: []openapi.EndpointSync{
				{
					Url:  ptr.To("10.0.0.1:32101"),
					Port: ptr.To(int32(32101)),
					SourceRef: &openapi.EndpointSyncAllOfSourceRef{
						Kind: ptr.To("Service"),
						Name: ptr.To("svc-nodeport"),
					},
				},
				{
					Url:  ptr.To("10.0.0.1:32102"),
					Port: ptr.To(int32(32102)),
					SourceRef: &openapi.EndpointSyncAllOfSourceRef{
						Kind: ptr.To("Service"),
						Name: ptr.To("svc-nodeport"),
					},
				},
				{
					Url:  ptr.To("10.0.0.1:31000"),
					Port: ptr.To(int32(31000)),
					SourceRef: &openapi.EndpointSyncAllOfSourceRef{
						Kind: ptr.To("Service"),
						Name: ptr.To("svc-clusterip"),
					},
				},
				{
					Url: ptr.To("https://ingress.example.com"),
					SourceRef: &openapi.EndpointSyncAllOfSourceRef{
						Kind: ptr.To("Ingress"),
						Name: ptr.To("svc-nodeport-ingress"),
					},
				},
			},
		}

		applyNetworkingDetailsToSyncFields(networkingStatus, syncFields)

		Expect(syncFields.NetworkStatus).NotTo(BeNil())
		Expect(syncFields.NetworkStatus.Services).To(HaveLen(2))
		nodePortService := syncFields.NetworkStatus.Services[0]
		Expect(nodePortService.GetName()).To(Equal("svc-nodeport"))
		Expect(nodePortService.Urls).To(HaveLen(2))
		Expect([]string{
			nodePortService.Urls[0].GetUrl(),
			nodePortService.Urls[1].GetUrl(),
		}).To(ConsistOf("10.0.0.1:32101", "10.0.0.1:32102"))
		Expect(nodePortService.Urls[0].GetSourceKind()).To(Equal("Service"))
		Expect(nodePortService.Urls[0].GetSourceName()).To(Equal("svc-nodeport"))
		Expect(syncFields.NetworkStatus.Services[1].Urls).To(BeEmpty())
		Expect(syncFields.Urls).To(Equal([]string{"https://custom.example.com"}))
		Expect(syncFields.Endpoints).To(HaveLen(4))
	})

	It("does not override an existing external URL", func() {
		networkingStatus := endpoints.NewNetworkingStatus(1)
		networkingStatus.Endpoints = []endpoints.EndpointStatus{
			{
				Services: []endpoints.ServiceStatus{
					{
						Name: "jupyter",
						URLs: []endpoints.ServiceURLStatus{
							{URL: "https://jupyter.example.com/lab"},
						},
					},
				},
			},
		}
		syncFields := &openapi.WorkloadSyncFields{
			Urls: []string{"https://user.example.com/custom"},
		}

		applyNetworkingDetailsToSyncFields(networkingStatus, syncFields)

		Expect(syncFields.Urls).To(Equal([]string{"https://user.example.com/custom"}))
		Expect(syncFields.NetworkStatus).NotTo(BeNil())
		Expect(syncFields.Endpoints).To(BeNil())
	})
})

var _ = Describe("parseStructureElements", func() {
	It("converts all fields from CRD to API format", func() {
		parentID := "parent-123"
		instanceKey := "frontend"
		replicaKey := "replica-0"
		se := &v1alpha1.StructureElements{
			Elements: []v1alpha1.StructureElement{
				{
					ID:          "elem-1",
					Name:        "deployment-worker",
					Level:       1,
					InstanceKey: &instanceKey,
				},
				{
					ID:         "elem-2",
					Name:       "replicaset-worker",
					Level:      2,
					ParentID:   &parentID,
					ReplicaKey: &replicaKey,
					K8sRef: &v1alpha1.K8sObjectRef{
						UID:  "uid-abc",
						Name: "worker-rs",
						GVK: v1alpha1.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "ReplicaSet",
						},
					},
				},
			},
		}

		result := parseStructureElements(se)
		Expect(result).To(HaveLen(2))

		// First element: level 1, no parent, no k8sRef
		Expect(result[0].Id).To(Equal("elem-1"))
		Expect(result[0].Name).To(Equal("deployment-worker"))
		Expect(result[0].HierarchyLevel).To(Equal(int32(1)))
		Expect(result[0].HasParentId()).To(BeFalse())
		Expect(result[0].HasInstanceKey()).To(BeTrue())
		Expect(result[0].GetInstanceKey()).To(Equal("frontend"))
		Expect(result[0].HasKubernetesRef()).To(BeFalse())

		// Second element: level 2, with parent, replicaKey, and k8sRef
		Expect(result[1].Id).To(Equal("elem-2"))
		Expect(result[1].HierarchyLevel).To(Equal(int32(2)))
		Expect(result[1].GetParentId()).To(Equal("parent-123"))
		Expect(result[1].GetReplicaKey()).To(Equal("replica-0"))
		Expect(result[1].HasKubernetesRef()).To(BeTrue())
		k8sRef := result[1].GetKubernetesRef()
		Expect(*k8sRef.Uid).To(Equal("uid-abc"))
		Expect(*k8sRef.Name).To(Equal("worker-rs"))
		Expect(*k8sRef.Gvk.Group).To(Equal("apps"))
		Expect(*k8sRef.Gvk.Version).To(Equal("v1"))
		Expect(*k8sRef.Gvk.Kind).To(Equal("ReplicaSet"))
	})

	It("returns empty slice for empty elements", func() {
		se := &v1alpha1.StructureElements{
			Elements: []v1alpha1.StructureElement{},
		}
		result := parseStructureElements(se)
		Expect(result).To(BeEmpty())
	})

	It("truncates to maxStructureElementsPerWorkload", func() {
		elements := make([]v1alpha1.StructureElement, maxStructureElementsPerWorkload+10)
		for i := range elements {
			elements[i] = v1alpha1.StructureElement{
				ID:    fmt.Sprintf("elem-%d", i),
				Name:  fmt.Sprintf("element-%d", i),
				Level: (i % 3) + 1,
			}
		}
		se := &v1alpha1.StructureElements{Elements: elements}

		result := parseStructureElements(se)
		Expect(result).To(HaveLen(maxStructureElementsPerWorkload))
		Expect(result[0].Id).To(Equal("elem-0"))
		Expect(result[maxStructureElementsPerWorkload-1].Id).To(Equal(fmt.Sprintf("elem-%d", maxStructureElementsPerWorkload-1)))
	})

	It("handles elements with no optional fields", func() {
		se := &v1alpha1.StructureElements{
			Elements: []v1alpha1.StructureElement{
				{ID: "elem-1", Name: "simple", Level: 1},
			},
		}
		result := parseStructureElements(se)
		Expect(result).To(HaveLen(1))
		Expect(result[0].Id).To(Equal("elem-1"))
		Expect(result[0].HasParentId()).To(BeFalse())
		Expect(result[0].HasInstanceKey()).To(BeFalse())
		Expect(result[0].HasReplicaKey()).To(BeFalse())
		Expect(result[0].HasKubernetesRef()).To(BeFalse())
	})
})

var _ = Describe("getExternalWorkloadAdditionalFields cliCommand", func() {
	const cliCommand = "runai workspace submit clitest -i ubuntu -g 1 --command -- sleep 30000"

	It("maps the CLI submit-command annotation to additionalFields.cliCommand (redirected CLI flow)", func() {
		annotations := map[string]string{commonconstants.WorkloadCliCommandAnnotation: cliCommand}

		result := TestGetExternalWorkloadAdditionalFields(v1alpha1.ExternalWorkloadSpec{}, nil, annotations)

		Expect(result).ToNot(BeNil())
		Expect(*result).To(MatchJSON(fmt.Sprintf(`{"cliCommand":%q}`, cliCommand)))
	})

	It("does not set cliCommand when the annotation is absent (native/non-CLI flow unchanged)", func() {
		result := TestGetExternalWorkloadAdditionalFields(v1alpha1.ExternalWorkloadSpec{}, nil, nil)

		Expect(result).To(BeNil())
	})

	It("preserves numeric replica fields from EWI", func() {
		raw := `{"statusReplicas":2,"readyReplicas":1,"keep":"value"}`
		result := TestGetExternalWorkloadAdditionalFields(
			v1alpha1.ExternalWorkloadSpec{AdditionalFields: &raw}, nil, nil)

		Expect(result).NotTo(BeNil())
		var additionalFields map[string]any
		Expect(json.Unmarshal([]byte(*result), &additionalFields)).To(Succeed())
		Expect(additionalFields).To(HaveKeyWithValue("statusReplicas", float64(2)))
		Expect(additionalFields).To(HaveKeyWithValue("readyReplicas", float64(1)))
		Expect(additionalFields).To(HaveKeyWithValue("keep", "value"))
	})

	It("ignores an empty CLI submit-command annotation", func() {
		annotations := map[string]string{commonconstants.WorkloadCliCommandAnnotation: ""}

		result := TestGetExternalWorkloadAdditionalFields(v1alpha1.ExternalWorkloadSpec{}, nil, annotations)

		Expect(result).To(BeNil())
	})

	It("preserves spec additionalFields and NIM flag alongside cliCommand", func() {
		existing := `{"existingField":"existingValue"}`
		spec := v1alpha1.ExternalWorkloadSpec{AdditionalFields: &existing}
		labels := map[string]string{constants.LabelNIMWorkload: "true"}
		annotations := map[string]string{commonconstants.WorkloadCliCommandAnnotation: cliCommand}

		result := TestGetExternalWorkloadAdditionalFields(spec, labels, annotations)

		Expect(result).ToNot(BeNil())
		Expect(*result).To(MatchJSON(fmt.Sprintf(
			`{"existingField":"existingValue","nimWorkload":"true","cliCommand":%q}`, cliCommand)))
	})

	It("adds vLLM marker only for the canonical true label", func() {
		additionalFields := map[string]any{}
		TestAddVLLMWorkloadFlagIfPresent(
			map[string]string{constants.LabelVLLMWorkload: "true"}, additionalFields)
		Expect(additionalFields).To(HaveKeyWithValue(constants.AdditionalVLLMWorkload, "true"))

		for _, value := range []string{"false", "TRUE", "", " true "} {
			additionalFields = map[string]any{}
			TestAddVLLMWorkloadFlagIfPresent(
				map[string]string{constants.LabelVLLMWorkload: value}, additionalFields)
			Expect(additionalFields).NotTo(HaveKey(constants.AdditionalVLLMWorkload))
		}
	})

	It("preserves existing fields with vLLM and NIM markers", func() {
		existing := `{"existingField":"existingValue"}`
		labels := map[string]string{
			constants.LabelNIMWorkload:  "true",
			constants.LabelVLLMWorkload: "true",
		}

		result := TestGetExternalWorkloadAdditionalFields(
			v1alpha1.ExternalWorkloadSpec{AdditionalFields: &existing}, labels, nil)

		Expect(result).NotTo(BeNil())
		Expect(*result).To(MatchJSON(
			`{"existingField":"existingValue","nimWorkload":"true","vllmWorkload":"true"}`))
	})
})
