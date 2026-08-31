/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package collector

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"go.uber.org/zap"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/diagnostics"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/progressbar/spinner"
	corev1 "k8s.io/api/core/v1"
)

const (
	concurrentLimit             = 50
	tmpLogFileNameFormat        = "%s_%s.log"
	tmpPrevPodLogFileNameFormat = "%s_%s_previous.log"
)

type CollectionReporterer interface {
	CreateReport() string
	ReportNamespace(ns string, err error)
	ReportContainer(ns, pod, container string, isPrev bool, err error)
}

type LogsClienter interface {
	ListPods(ctx context.Context, namespace string) (*corev1.PodList, error)
	GetPodLogStream(ctx context.Context, podName string, namespace string, opts *corev1.PodLogOptions) (io.ReadCloser, error)
}

type ArchiveAppender interface {
	AppendFile(srcPath, dstPath string) error
}

type LogsCollector struct {
	logsClient        LogsClienter
	archiveAppender   ArchiveAppender
	concurrentLimitCh chan struct{}
	noPrevious        bool
	outputDirPath     string
}

func NewLogsCollector(logsClient LogsClienter, archiveAppender ArchiveAppender, outputDirPath string, noPrevious bool) *LogsCollector {
	return &LogsCollector{
		logsClient:        logsClient,
		archiveAppender:   archiveAppender,
		concurrentLimitCh: make(chan struct{}, concurrentLimit),
		noPrevious:        noPrevious,
		outputDirPath:     outputDirPath,
	}
}

func (logsCollector *LogsCollector) Collect(ctx context.Context, namespaces []string) {
	collectionReporter := NewCollectionReporter(logsCollector.noPrevious)

	sp := spinner.NewSpinner(diagnostics.CollectLogsSpinnerMsg)
	go sp.Run()
	logsCollector.collect(ctx, collectionReporter, namespaces)
	sp.Stop()
	config.GetInstance().GetLogger().Info(collectionReporter.CreateReport())
}

func (logsCollector *LogsCollector) collect(ctx context.Context, collectionReporter CollectionReporterer, namespaces []string) {
	wg := sync.WaitGroup{}
	for _, ns := range namespaces {
		podList, err := logsCollector.logsClient.ListPods(ctx, ns)
		collectionReporter.ReportNamespace(ns, err)
		if err != nil || podList == nil {
			continue
		}

		logsCollector.collectNamespaceLogs(ctx, collectionReporter, podList, &wg, ns)
	}
	wg.Wait()
}

func (logsCollector *LogsCollector) collectNamespaceLogs(ctx context.Context, collectionReporter CollectionReporterer, podList *corev1.PodList, wg *sync.WaitGroup, ns string) {
	for _, pod := range podList.Items {
		wg.Add(1)
		select {
		case <-ctx.Done():
			wg.Done()
			config.GetInstance().GetLogger().Info("context canceled")
			return
		case logsCollector.concurrentLimitCh <- struct{}{}:
			go func(pod *corev1.Pod) {
				defer func() {
					wg.Done()
					<-logsCollector.concurrentLimitCh
				}()
				logsCollector.collectPodLogs(ctx, collectionReporter, ns, pod)
			}(&pod)
		}
	}
}

func (logsCollector *LogsCollector) collectPodLogs(ctx context.Context, collectionReporter CollectionReporterer, ns string, pod *corev1.Pod) {
	prevOpts := logsCollector.getPodCollectLogsOptions()
	for _, container := range pod.Spec.Containers {
		for _, prev := range prevOpts {
			exists, err := logsCollector.collectContainerLogs(ctx, container.Name, pod, prev)
			if !exists {
				continue
			}
			collectionReporter.ReportContainer(ns, pod.Name, container.Name, prev, err)
		}
	}
}

func (logsCollector *LogsCollector) collectContainerLogs(ctx context.Context, containerName string, pod *corev1.Pod, prev bool) (bool, error) {
	podLogsStream, err := logsCollector.logsClient.GetPodLogStream(ctx, pod.Name, pod.Namespace, &corev1.PodLogOptions{
		Container: containerName,
		Previous:  prev,
	})
	if err != nil {
		if logsCollector.isNotFoundErr(err, prev) {
			return false, nil
		}
		return true, err
	}

	tmpLogFileName := fmt.Sprintf(tmpLogFileNameFormat, pod.Name, containerName)
	if prev {
		tmpLogFileName = fmt.Sprintf(tmpPrevPodLogFileNameFormat, pod.Name, containerName)
	}

	archiveFilePath := NewPodInfo(pod).getLogFilePathInArchive(containerName, prev)
	tmpLogFilePath := filepath.Join(logsCollector.outputDirPath, tmpLogFileName)
	return true, logsCollector.archiveLogs(archiveFilePath, tmpLogFilePath, podLogsStream)
}

func (logsCollector *LogsCollector) archiveLogs(archiveFilePath, tmpLogFilePath string, podLogsStream io.ReadCloser) error {
	tmpLogsFile, err := os.Create(tmpLogFilePath)
	if err != nil {
		return err
	}

	defer func() {
		if err = tmpLogsFile.Close(); err != nil {
			config.GetInstance().GetLogger().Error("failed to close log file ", zap.Error(err))
		}
		if err = os.Remove(tmpLogFilePath); err != nil {
			config.GetInstance().GetLogger().Error("failed to delete created log file", zap.Error(err))
		}
	}()

	if _, err = io.Copy(tmpLogsFile, podLogsStream); err != nil {
		return err
	}

	return logsCollector.archiveAppender.AppendFile(tmpLogFilePath, archiveFilePath)
}

func (logsCollector *LogsCollector) getPodCollectLogsOptions() []bool {
	if logsCollector.noPrevious {
		return []bool{false}
	}
	return []bool{false, true}
}

func (logsCollector *LogsCollector) isNotFoundErr(err error, prev bool) bool {
	return apierrors.IsNotFound(err) || (apierrors.IsBadRequest(err) && prev && strings.Contains(err.Error(), "not found"))
}
