/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package collector

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	collectormocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/diagnostics/collector"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Collector", func() {
	var pod corev1.Pod
	ownerName := "runai-backend-workloads-b8f4ff97d"
	namespace, containerName, podName, nodeName :=
		"runai-backend",
		"nginx-container",
		"runai-backend-workloads-b8f4ff97d-49l72",
		"control-plane-1"

	Describe("Collect", func() {
		var mockCtrl *gomock.Controller
		var ctx context.Context

		BeforeEach(func() {
			ctx = context.Background()
			mockCtrl = gomock.NewController(GinkgoT())
			pod = corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{{
						Name: ownerName, Kind: string(ReplicaSet),
					}},
				},
				Spec: corev1.PodSpec{
					NodeName: nodeName, Containers: []corev1.Container{{
						Name: containerName,
					}}},
			}
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		It("should not call mocks when no namespaces provided", func() {
			lc := NewLogsCollector(nil, nil, "./", true)
			lc.collect(ctx, nil, []string{})
		})

		It("should handle failure to list pods, should report", func() {
			archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
			logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
			reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)
			logsClient.EXPECT().ListPods(ctx, namespace).Return(nil, fmt.Errorf("failed"))
			reportCollector.EXPECT().ReportNamespace(namespace, fmt.Errorf("failed"))
			lc := NewLogsCollector(logsClient, archiveAppender, "./", true)
			lc.collect(ctx, reportCollector, []string{namespace})
		})

		It("no pods, should report", func() {
			archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
			logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
			reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)
			logsClient.EXPECT().ListPods(ctx, namespace).Return(nil, nil)
			reportCollector.EXPECT().ReportNamespace(namespace, nil)
			lc := NewLogsCollector(logsClient, archiveAppender, "./", true)
			lc.collect(ctx, reportCollector, []string{namespace})
		})

		Context("with pods", func() {
			Context("no previous", func() {
				It("happy flow", func() {
					archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
					logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
					reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)

					logsClient.EXPECT().ListPods(ctx, namespace).Return(&corev1.PodList{Items: []corev1.Pod{pod}}, nil)
					reportCollector.EXPECT().ReportNamespace(namespace, nil)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  false,
					}).Return(newTestReadCloser("first log"), nil).Times(1)
					p := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
					archiveAppender.EXPECT().AppendFile(fmt.Sprintf("%s_%s.log", podName, containerName), p).Return(nil).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, false, nil).Times(1)
					lc := NewLogsCollector(logsClient, archiveAppender, "./", true)
					lc.collect(ctx, reportCollector, []string{namespace})
				})

				It("fail to get pod logs, should report", func() {
					archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
					logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
					reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)
					logsClient.EXPECT().ListPods(ctx, namespace).Return(&corev1.PodList{Items: []corev1.Pod{pod}}, nil)
					reportCollector.EXPECT().ReportNamespace(namespace, nil)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  false,
					}).Return(nil, fmt.Errorf("failed to get pod logs")).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, false, fmt.Errorf("failed to get pod logs")).Times(1)
					lc := NewLogsCollector(logsClient, archiveAppender, "./", true)
					lc.collect(ctx, reportCollector, []string{namespace})
				})

				It("fail to append pod logs file to tar, should report", func() {
					archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
					logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
					reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)

					logsClient.EXPECT().ListPods(ctx, namespace).Return(&corev1.PodList{Items: []corev1.Pod{pod}}, nil)
					reportCollector.EXPECT().ReportNamespace(namespace, nil)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  false,
					}).Return(newTestReadCloser("log stram"), nil).Times(1)
					p := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
					archiveAppender.EXPECT().AppendFile(fmt.Sprintf("%s_%s.log", podName, containerName), p).Return(fmt.Errorf("failed to append file")).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, false, fmt.Errorf("failed to append file")).Times(1)
					lc := NewLogsCollector(logsClient, archiveAppender, "./", true)
					lc.collect(ctx, reportCollector, []string{namespace})
				})
			})

			Context("with previous", func() {
				It("happy flow", func() {
					archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
					logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
					reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)

					logsClient.EXPECT().ListPods(ctx, namespace).Return(&corev1.PodList{Items: []corev1.Pod{pod}}, nil)
					reportCollector.EXPECT().ReportNamespace(namespace, nil)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  false,
					}).Return(newTestReadCloser("log"), nil).Times(1)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  true,
					}).Return(newTestReadCloser("previous pod log"), nil).Times(1)
					p := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
					archiveAppender.EXPECT().AppendFile(fmt.Sprintf("%s_%s.log", podName, containerName), p).Return(nil).Times(1)
					p = NewPodInfo(&pod).getLogFilePathInArchive(containerName, true)
					archiveAppender.EXPECT().AppendFile(fmt.Sprintf("%s_%s_previous.log", podName, containerName), p).Return(nil).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, false, nil).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, true, nil).Times(1)
					lc := NewLogsCollector(logsClient, archiveAppender, "./", false)
					lc.collect(ctx, reportCollector, []string{namespace})
				})

				It("one container fail to get logs, should report both", func() {
					archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
					logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
					reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)

					logsClient.EXPECT().ListPods(ctx, namespace).Return(&corev1.PodList{Items: []corev1.Pod{pod}}, nil)
					reportCollector.EXPECT().ReportNamespace(namespace, nil)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  false,
					}).Return(nil, fmt.Errorf("failed to got pod logs")).Times(1)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  true,
					}).Return(newTestReadCloser("first log"), nil).Times(1)
					p := NewPodInfo(&pod).getLogFilePathInArchive(containerName, true)
					archiveAppender.EXPECT().AppendFile(fmt.Sprintf("%s_%s_previous.log", podName, containerName), p).Return(nil).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, false, fmt.Errorf("failed to got pod logs")).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, true, nil).Times(1)
					lc := NewLogsCollector(logsClient, archiveAppender, "./", false)
					lc.collect(ctx, reportCollector, []string{namespace})

				})

				It("one container fail to append log file, should report both", func() {
					archiveAppender := collectormocks.NewMockArchiveAppender(mockCtrl)
					logsClient := collectormocks.NewMockLogsClienter(mockCtrl)
					reportCollector := collectormocks.NewMockCollectionReporterer(mockCtrl)

					logsClient.EXPECT().ListPods(ctx, namespace).Return(&corev1.PodList{Items: []corev1.Pod{pod}}, nil)
					reportCollector.EXPECT().ReportNamespace(namespace, nil)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  false,
					}).Return(newTestReadCloser("log"), nil).Times(1)
					logsClient.EXPECT().GetPodLogStream(ctx, podName, namespace, &corev1.PodLogOptions{
						Container: containerName,
						Previous:  true,
					}).Return(newTestReadCloser("prev pod log"), nil).Times(1)
					p := NewPodInfo(&pod).getLogFilePathInArchive(containerName, false)
					archiveAppender.EXPECT().AppendFile(fmt.Sprintf("%s_%s.log", podName, containerName), p).Return(nil).Times(1)
					p = NewPodInfo(&pod).getLogFilePathInArchive(containerName, true)
					archiveAppender.EXPECT().AppendFile(fmt.Sprintf("%s_%s_previous.log", podName, containerName), p).Return(fmt.Errorf("failed to open file")).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, true, fmt.Errorf("failed to open file")).Times(1)
					reportCollector.EXPECT().ReportContainer(namespace, podName, containerName, false, nil).Times(1)
					lc := NewLogsCollector(logsClient, archiveAppender, "./", false)
					lc.collect(ctx, reportCollector, []string{namespace})

				})
			})
		})
	})
})

func newTestReadCloser(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
