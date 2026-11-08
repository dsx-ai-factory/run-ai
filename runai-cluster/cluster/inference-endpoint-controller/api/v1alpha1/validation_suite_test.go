/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/yaml"
)

var (
	testEnv   *envtest.Environment
	k8sClient client.Client
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InferenceEndpoint API Suite")
}

var _ = BeforeSuite(func() {
	Expect(AddToScheme(scheme.Scheme)).To(Succeed())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	if os.Getenv("KUBEBUILDER_ASSETS") == "" {
		testEnv.BinaryAssetsDirectory = findEnvtestBinaries()
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(testEnv.Stop()).To(Succeed())
})

// findEnvtestBinaries returns the directory under bin/k8s that holds
// kube-apiserver, as installed by `make setup-envtest`.
func findEnvtestBinaries() string {
	var dir string
	_ = filepath.WalkDir(filepath.Join("..", "..", "bin", "k8s"), func(path string, d fs.DirEntry, err error) error {
		if err == nil && dir == "" && !d.IsDir() && d.Name() == "kube-apiserver" {
			dir = filepath.Dir(path)
		}
		return nil
	})
	return dir
}

func validEndpoint() *InferenceEndpoint {
	return &InferenceEndpoint{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "ie-", Namespace: "default"},
		Spec: InferenceEndpointSpec{
			Workload: WorkloadSpec{
				Selector: WorkloadSelector{MatchLabels: map[string]string{"app": "llama"}},
				Port:     8000,
			},
			Routing: RoutingSpec{
				GatewayRef: GatewayReference{Name: "runai-gateway"},
				Hostnames:  []string{"llama.example.com"},
			},
			Epp: EppSpec{Profile: EppProfileGIEV15},
		},
	}
}

var _ = Describe("InferenceEndpoint CRD validation", func() {
	ctx := context.Background()

	It("accepts a minimal endpoint and defaults the path prefix to /v1", func() {
		ie := validEndpoint()
		Expect(k8sClient.Create(ctx, ie)).To(Succeed())
		Expect(ie.Spec.Routing.PathPrefix).To(Equal("/v1"))
		Expect(ie.Spec.Epp.Scheduling).To(BeNil())
	})

	It("keeps an explicit path prefix", func() {
		ie := validEndpoint()
		ie.Spec.Routing.PathPrefix = "/openai/v1"
		Expect(k8sClient.Create(ctx, ie)).To(Succeed())
		Expect(ie.Spec.Routing.PathPrefix).To(Equal("/openai/v1"))
	})

	It("accepts a wildcard hostname", func() {
		ie := validEndpoint()
		ie.Spec.Routing.Hostnames = []string{"*.models.example.com", "llama"}
		Expect(k8sClient.Create(ctx, ie)).To(Succeed())
	})

	It("accepts scheduling scorers", func() {
		ie := validEndpoint()
		ie.Spec.Epp.Scheduling = &SchedulingSpec{Scorers: []ScorerSpec{
			{Type: ScorerQueue, Weight: 30},
			{Type: ScorerKVCache, Weight: 20},
			{Type: ScorerPrefixCache, Weight: 0},
		}}
		Expect(k8sClient.Create(ctx, ie)).To(Succeed())
	})

	It("accepts every sample in config/samples", func() {
		files, err := filepath.Glob(filepath.Join("..", "..", "config", "samples", "inference_v1alpha1_inferenceendpoint*.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(files).NotTo(BeEmpty())
		for _, file := range files {
			data, err := os.ReadFile(file)
			Expect(err).NotTo(HaveOccurred())
			sample := &unstructured.Unstructured{}
			Expect(yaml.Unmarshal(data, &sample.Object)).To(Succeed(), file)

			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: sample.GetNamespace()}}
			Expect(client.IgnoreAlreadyExists(k8sClient.Create(ctx, ns))).To(Succeed(), file)
			// Dry run: several samples share the name llama3.
			err = k8sClient.Create(ctx, sample, client.DryRunAll)
			Expect(err).NotTo(HaveOccurred(), "%s: %v", file, err)
			Expect(apierrors.IsInvalid(err)).To(BeFalse(), file)
		}
	})

	DescribeTable("rejects an invalid spec",
		func(mutate func(*InferenceEndpoint)) {
			ie := validEndpoint()
			mutate(ie)
			Expect(k8sClient.Create(ctx, ie)).NotTo(Succeed())
		},
		Entry("empty workload selector", func(ie *InferenceEndpoint) {
			ie.Spec.Workload.Selector.MatchLabels = map[string]string{}
		}),
		Entry("port 0", func(ie *InferenceEndpoint) { ie.Spec.Workload.Port = 0 }),
		Entry("port above 65535", func(ie *InferenceEndpoint) { ie.Spec.Workload.Port = 65536 }),
		Entry("no gateway name", func(ie *InferenceEndpoint) { ie.Spec.Routing.GatewayRef.Name = "" }),
		Entry("no hostnames", func(ie *InferenceEndpoint) { ie.Spec.Routing.Hostnames = nil }),
		Entry("hostname with a space", func(ie *InferenceEndpoint) { ie.Spec.Routing.Hostnames = []string{"foo bar"} }),
		Entry("upper case hostname", func(ie *InferenceEndpoint) { ie.Spec.Routing.Hostnames = []string{"Llama.example.com"} }),
		Entry("hostname longer than 253 characters", func(ie *InferenceEndpoint) {
			ie.Spec.Routing.Hostnames = []string{strings.Repeat("a", 254)}
		}),
		Entry("more than 16 hostnames", func(ie *InferenceEndpoint) {
			for i := 0; i < 17; i++ {
				ie.Spec.Routing.Hostnames = append(ie.Spec.Routing.Hostnames, fmt.Sprintf("h%d.example.com", i))
			}
		}),
		Entry("path prefix without a leading slash", func(ie *InferenceEndpoint) {
			ie.Spec.Routing.PathPrefix = "v1"
		}),
		Entry("no EPP profile", func(ie *InferenceEndpoint) { ie.Spec.Epp.Profile = "" }),
		Entry("unsupported EPP profile", func(ie *InferenceEndpoint) { ie.Spec.Epp.Profile = "llm-d-v0.10" }),
		Entry("unknown scorer type", func(ie *InferenceEndpoint) {
			ie.Spec.Epp.Scheduling = &SchedulingSpec{Scorers: []ScorerSpec{{Type: "random", Weight: 1}}}
		}),
		Entry("scorer weight above 100", func(ie *InferenceEndpoint) {
			ie.Spec.Epp.Scheduling = &SchedulingSpec{Scorers: []ScorerSpec{{Type: ScorerQueue, Weight: 101}}}
		}),
		Entry("duplicate scorer type", func(ie *InferenceEndpoint) {
			ie.Spec.Epp.Scheduling = &SchedulingSpec{Scorers: []ScorerSpec{
				{Type: ScorerQueue, Weight: 1}, {Type: ScorerQueue, Weight: 2},
			}}
		}),
	)
})
