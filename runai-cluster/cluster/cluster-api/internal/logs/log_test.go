/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	namespace   = "runai-test"
	testPodName = "test-pod"
)

var _ = Describe("streamer", func() {
	var (
		streamer   *PodStreamer
		fakeclient *fake.Clientset
	)
	BeforeEach(func() {
		fakeclient = fake.NewSimpleClientset()
		streamer = NewPodStreamer(fakeclient)
	})
	Context("NewPodStreamer", func() {
		It("should return a new PodStreamer", func() {
			Expect(streamer).NotTo(BeNil())
		})
		It("should return a PodStreamer with the correct client", func() {
			Expect(streamer.client).To(Equal(fakeclient))
		})
	})
	Context("Stream", func() {
		var pod *v1.Pod
		BeforeEach(func() {
			// create pod
			pod = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testPodName,
					Namespace: namespace,
				},
			}
			_, err := fakeclient.CoreV1().Pods(namespace).Create(context.Background(),
				pod, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return a stream", func() {
			stream, err := streamer.Stream(context.Background(), namespace, &Options{Pod: testPodName})
			Expect(err).NotTo(HaveOccurred())
			Expect(stream).NotTo(BeNil())
		})
	})
})
