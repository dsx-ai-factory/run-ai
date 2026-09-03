/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package collector

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"golang.org/x/exp/slices"

	"golang.org/x/exp/maps"
)

type ContainerReport struct {
	Container string
	Pod       string
	Namespace string
	Err       error
	Previous  bool
}

type NamespaceReport struct {
	listErr          error
	SuccessCount     int
	FailCount        int
	ContainerReports []*ContainerReport
}

type CollectionReporter struct {
	reportByNamespace map[string]*NamespaceReport
	noPrevious        bool
	mu                sync.Mutex
}

func NewCollectionReporter(noPrevious bool) *CollectionReporter {
	return &CollectionReporter{
		reportByNamespace: make(map[string]*NamespaceReport),
		noPrevious:        noPrevious,
		mu:                sync.Mutex{},
	}
}

func (cr *CollectionReporter) CreateReport() string {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	tableFormatter := newReportSummaryPrinter(cr.reportByNamespace, cr.noPrevious)
	failedNamespaces := cr.getFailedNamespacesLogs()

	report := &strings.Builder{}
	report.WriteString("Summary:\n")
	writeSection(report, "Namespaces:", strings.Join(failedNamespaces, "\n"))
	writeSection(report, tableFormatter.Subject(), tableFormatter.SummaryTable())
	writeSection(report, "", cr.getFailedPodsByNamespaces())
	return report.String()
}

func (cr *CollectionReporter) getFailedPodsByNamespaces() string {
	strBuilder := &strings.Builder{}
	failedNamespacePodsMap := cr.getFailedPodsLogs()
	sortedNamespaces := maps.Keys(failedNamespacePodsMap)
	slices.Sort(sortedNamespaces)
	for _, namespace := range sortedNamespaces {
		failedPods := failedNamespacePodsMap[namespace]
		writeSection(strBuilder, fmt.Sprintf("%s:", namespace), failedPods)
	}

	return strBuilder.String()
}

func (cr *CollectionReporter) ReportNamespace(ns string, err error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.reportByNamespace[ns] == nil {
		cr.reportByNamespace[ns] = &NamespaceReport{ContainerReports: []*ContainerReport{}}
	}

	cr.reportByNamespace[ns].listErr = err
}

func (cr *CollectionReporter) ReportContainer(ns, pod, container string, isPrev bool, err error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.reportByNamespace[ns] == nil {
		cr.reportByNamespace[ns] = &NamespaceReport{ContainerReports: []*ContainerReport{}}
	}

	nr := cr.reportByNamespace[ns]
	nr.ContainerReports = append(nr.ContainerReports, &ContainerReport{Namespace: ns, Pod: pod, Container: container, Previous: isPrev, Err: err})
	if err == nil {
		nr.SuccessCount += 1
	} else {
		nr.FailCount += 1
	}
}

func (cr *CollectionReporter) getFailedNamespacesLogs() (failedNamespaces []string) {
	namespaces := maps.Keys(cr.reportByNamespace)
	slices.Sort(namespaces)
	for _, namespace := range namespaces {
		if cr.reportByNamespace[namespace].listErr != nil {
			failedNamespaces = append(failedNamespaces, fmt.Sprintf("Failed to list %s pods: [%+v]", namespace, cr.reportByNamespace[namespace].listErr))
		}
	}

	return failedNamespaces
}

func (cr *CollectionReporter) getFailedPodsLogs() map[string]string {
	failedPods := make(map[string]string)
	for namespace := range cr.reportByNamespace {
		if cr.reportByNamespace[namespace].listErr != nil {
			continue
		}

		failedToListPods := cr.getPodsErrors(namespace)
		if len(failedToListPods) > 0 {
			failedPods[namespace] = strings.Join(failedToListPods, "\n")
		}
	}

	return failedPods
}

func (cr *CollectionReporter) getPodsErrors(namespace string) (errs []string) {
	namespaceContainersReports := cr.reportByNamespace[namespace].ContainerReports
	sort.Slice(namespaceContainersReports, func(i, j int) bool {
		return namespaceContainersReports[i].Pod < namespaceContainersReports[j].Pod
	})

	for _, f := range namespaceContainersReports {
		if f.Err != nil {
			errs = append(errs, fmt.Sprintf(getPodErrorFormat(f), f.Pod, f.Container, f.Err))
		}
	}
	return errs
}

func getPodErrorFormat(f *ContainerReport) string {
	if f.Previous {
		return "Failed to get logs of %s/%s (previous): [%+v]"
	}
	return "Failed to get logs of %s/%s: [%+v]"
}

func writeSection(report *strings.Builder, subject, data string) {
	if len(data) == 0 {
		return
	}

	if len(subject) > 0 {
		report.WriteByte('\n')
		report.WriteString(subject)
	}
	report.WriteByte('\n')
	report.WriteString(data)
	report.WriteByte('\n')
}
