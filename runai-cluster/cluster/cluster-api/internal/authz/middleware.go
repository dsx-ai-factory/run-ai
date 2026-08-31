/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authz

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rs/zerolog/log"
	authzapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/auth"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz/authorizers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/project"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
)

const (
	workloadIDAnnotationKey = "workloadId"
	unauthorizedMessage     = "User %s is not authorized to perform the %s" +
		" action on the %s workload named %s in the project %s."
)

type Middleware struct {
	authorizer authorizers.Authorizer
	locator    workload.Locator
}

func NewMiddleware(authorizer authorizers.Authorizer, locator workload.Locator) *Middleware {
	log.Debug().Msg("Creating new authorization middleware")

	middleware := &Middleware{
		authorizer: authorizer,
		locator:    locator,
	}

	log.Debug().Msg("Authorization middleware created successfully")

	return middleware
}

func (m *Middleware) Middleware(next http.Handler) http.Handler {
	log.Debug().Msg("Creating middleware handler")
	return m.middleware(next)
}

func (m *Middleware) middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("method", r.Method).Str("url", r.URL.String()).Msg("Starting request processing")

		ctx := r.Context()

		request, err := createRequest(ctx, m.locator)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create request from context")
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		log.Debug().Interface("request", request).Msg("Request created from context")

		ctx = context.WithValue(ctx, authzapi.ContextAccessToken, request.AuthInfo.RawToken)

		if _, reason, err := m.authorizer.Authorize(ctx, request); err != nil {
			if reason == authorizers.ReasonUnauthorized {
				log.Warn().Err(err).Msg("Forbidden request")
				http.Error(w, createUnauthorizedMessage(request), http.StatusForbidden)

				return
			}

			log.Error().Err(err).Msg("Authorization failed")
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		next.ServeHTTP(w, r)
	}
}

func createUnauthorizedMessage(request *authorizers.Request) string {
	if request == nil {
		log.Error().Msg("Request is nil")
		return "unauthorized"
	}

	message := fmt.Sprintf(unauthorizedMessage, request.AuthInfo.Username, request.Action,
		request.Kind, request.Workload, request.Project)
	log.Debug().Str("message", message).Msg("Created unauthorized message")

	return message
}

func createRequest(ctx context.Context, locator workload.Locator) (*authorizers.Request, error) {
	log.Debug().Msg("Starting to create request from context")

	namespace := project.NamespaceFromContext(ctx)
	if namespace == "" {
		log.Error().Msg("Namespace not found in context")
		return nil, errors.New("namespace is required, but not found in context")
	}

	authInfo := auth.FromContext(ctx)
	if authInfo == nil {
		log.Error().Msg("Auth info not found in context")
		return nil, errors.New("auth info is required, but not found in context")
	}

	action, err := api.WorkloadActionFromContext(ctx)
	if err != nil || action == -1 {
		log.Error().Msg("Workload action not found in context")
		return nil, errors.New("workload action is required, but not found in context")
	}

	workloadName := api.WorkloadNameFromContext(ctx)
	if workloadName == "" {
		log.Error().Msg("Workload name not found in context")
		return nil, errors.New("workload name is required, but not found in context")
	}

	workloadType, err := api.WorkloadTypeFromContext(ctx)
	if err != nil || workloadType == "" {
		log.Error().Msg("Workload type not found in context")
		return nil, errors.New("workload type is required, but not found in context")
	}

	projectName := project.FromContext(ctx)
	if projectName == "" {
		log.Error().Msg("Project name not found in context")
		return nil, errors.New("project name is required, but not found in context")
	}

	workloadResource, err := locator.Locate(ctx, workloadType, workloadName, namespace)
	if err != nil {
		log.Error().Msg("could not locate desired workload in the cluster")
		return nil, errors.New("could not locate desired workload in the cluster")
	}

	workloadUID := GetWorkloadUID(workloadResource, workloadType)

	request := &authorizers.Request{
		AuthInfo:    authInfo,
		Namespace:   namespace,
		Action:      action,
		Workload:    workloadName,
		Kind:        workloadType,
		Project:     projectName,
		WorkloadUID: workloadUID,
	}

	log.Debug().Interface("request", request).Msg("Request created successfully")

	return request, nil
}

func GetWorkloadUID(resource client.Object, workloadType api.WorkloadType) string {
	annotations := resource.GetAnnotations()
	if workloadType != api.External {
		return annotations[workloadIDAnnotationKey]
	}

	// Try to get the Run:AI workload ID from the annotation
	// This is the preferred identifier for authorization
	if annotations != nil {
		if workloadID := annotations[workloadIDAnnotationKey]; workloadID != "" {
			return workloadID
		}
	}

	externalWorkload, ok := resource.(*v1alpha1.ExternalWorkload)
	if !ok {
		return ""
	}

	// Fallback to Kubernetes UID if the annotation is not present
	return string(externalWorkload.Spec.OriginalWorkloadUID)
}
