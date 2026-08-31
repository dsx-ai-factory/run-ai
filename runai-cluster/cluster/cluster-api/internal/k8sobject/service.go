/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8sobject

import (
	"context"
	"errors"
	"net/http"
	"time"

	openapi "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/logging"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	k8sGetTimeout   = 10 * time.Second
	upstreamTimeout = 10 * time.Second
)

// Service requires a non-cached k8s client — cached clients need scheme
// registration per kind, which would defeat the "any kind" intent.
type Service struct {
	nonCachedClient client.Client
	elements        ElementGetter
	pods            PodGetter
}

func NewService(nonCachedClient client.Client, elements ElementGetter, pods PodGetter) *Service {
	return &Service{
		nonCachedClient: nonCachedClient,
		elements:        elements,
		pods:            pods,
	}
}

func (s *Service) GetWorkloadObject(ctx context.Context, _ string) (openapi.ImplResponse, error) {
	workloadInfo, errResp, err := requireWorkloadInfo(ctx, "GetWorkloadObject")
	if err != nil {
		return errResp, err
	}

	gvk, err := workloadGVK(workloadInfo)
	if err != nil {
		logging.Errorf("k8sobject: GVK resolution failed for workload %s/%s: %v", workloadInfo.Namespace, workloadInfo.Name, err)
		return openapi.Response(http.StatusInternalServerError, nil), err
	}

	return s.fetchAndWrap(gvk, workloadInfo.Namespace, workloadInfo.Name)
}

func (s *Service) GetPodObject(ctx context.Context, workloadID string, podID string) (openapi.ImplResponse, error) {
	workloadInfo, errResp, err := requireWorkloadInfo(ctx, "GetPodObject")
	if err != nil {
		return errResp, err
	}

	upstreamCtx, cancel := context.WithTimeout(ctx, upstreamTimeout)
	defer cancel()

	pod, httpResp, err := s.pods.GetWorkloadPod(upstreamCtx, workloadID, podID)
	if httpResp != nil && httpResp.Body != nil {
		defer func() { _ = httpResp.Body.Close() }()
	}
	if httpResp == nil || httpResp.StatusCode != http.StatusOK {
		return upstreamErrorResponse("GetWorkloadPod", workloadID, podID, httpResp, err)
	}

	return s.fetchAndWrap(podGVK, workloadInfo.Namespace, pod.Name)
}

func (s *Service) GetElementObject(ctx context.Context, workloadID string, elementID string) (openapi.ImplResponse, error) {
	workloadInfo, errResp, err := requireWorkloadInfo(ctx, "GetElementObject")
	if err != nil {
		return errResp, err
	}

	upstreamCtx, cancel := context.WithTimeout(ctx, upstreamTimeout)
	defer cancel()

	elem, httpResp, err := s.elements.GetWorkloadElement(upstreamCtx, workloadID, elementID)
	if httpResp != nil && httpResp.Body != nil {
		defer func() { _ = httpResp.Body.Close() }()
	}
	if httpResp == nil || httpResp.StatusCode != http.StatusOK {
		return upstreamErrorResponse("GetWorkloadElement", workloadID, elementID, httpResp, err)
	}

	ref, refOK := elem.GetKubernetesRefOk()
	if !refOK || ref == nil {
		return openapi.Response(http.StatusNotFound, nil), nil
	}
	gvk, name, ok := elementGvkName(ref)
	if !ok {
		return openapi.Response(http.StatusNotFound, nil), nil
	}

	return s.fetchAndWrap(gvk, workloadInfo.Namespace, name)
}

// k8s GETs use context.Background() because controller-runtime authenticates
// as cluster-api's ServiceAccount — the user's ctx is irrelevant here.
// Status mapping is deliberately coarse: IsForbidden means cluster-api SA is
// mis-RBAC'd (server bug, not caller's), IsTimeout/IsTooManyRequests are
// transient — 500 + retry covers both.
func (s *Service) fetchAndWrap(gvk schema.GroupVersionKind, namespace, name string) (openapi.ImplResponse, error) {
	// Empty namespace/name fail k8s client validation (not IsNotFound) and would 500; report not-found instead.
	if namespace == "" || name == "" {
		return openapi.Response(http.StatusNotFound, nil), nil
	}

	u := unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)

	k8sCtx, cancel := context.WithTimeout(context.Background(), k8sGetTimeout)
	defer cancel()

	err := s.nonCachedClient.Get(k8sCtx, client.ObjectKey{Namespace: namespace, Name: name}, &u)
	retrievedAt := time.Now().UTC()
	if err != nil {
		if apierrors.IsNotFound(err) {
			return openapi.Response(http.StatusNotFound, nil), nil
		}
		logging.Errorf("k8sobject: get <%s> <%s/%s>: %v", gvk.String(), namespace, name, err)
		return openapi.Response(http.StatusInternalServerError, nil), err
	}

	sanitize(&u)
	return openapi.Response(http.StatusOK, openapi.Object{
		Object:      u.Object,
		RetrievedAt: retrievedAt,
	}), nil
}

// upstreamErrorResponse maps a failed workloads-client call to an ImplResponse.
// Passes the upstream status code through when present; falls back to 500 on
// transport failure (httpResp == nil).
func upstreamErrorResponse(op, workloadID, id string, httpResp *http.Response, err error) (openapi.ImplResponse, error) {
	status := http.StatusInternalServerError
	if httpResp != nil {
		status = httpResp.StatusCode
	}
	logging.Errorf("k8sobject: %s(%s/%s): status=%d, err=%v", op, workloadID, id, status, err)
	return openapi.Response(status, nil), err
}

// requireWorkloadInfo unwraps the middleware-populated workload from ctx.
// Missing context is a server-side misconfiguration (not a client error), so
// it surfaces as 500 with a single error log.
func requireWorkloadInfo(ctx context.Context, handler string) (*AuthorizedWorkloadInfo, openapi.ImplResponse, error) {
	workloadInfo, ok := WorkloadInfoFromContext(ctx)
	if ok {
		return workloadInfo, openapi.ImplResponse{}, nil
	}
	logging.Errorf("k8sobject: %s: workload not in context (middleware misconfigured)", handler)
	return nil, openapi.Response(http.StatusInternalServerError, nil), errors.New("workload not in context (middleware misconfigured)")
}
