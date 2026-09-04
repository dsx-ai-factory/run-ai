/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kartacontroller

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/kartacontroller/handlers"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	rateLimiterBaseDelay = 500 * time.Millisecond
	rateLimiterMaxDelay  = 60 * time.Second
)

type KartaController struct {
	Client client.Client

	rbacHandler              *handlers.KartaRBACHandler
	registeredAdditionalGVKs map[schema.GroupVersionKind]bool
	deletionHandler          *handlers.KartaDeletionHandler

	rbacUpdaterEnabled  bool
	clusterRoleReadName string
	clusterRoleCrudName string
}

type KartaControlParams struct {
	RBACUpdaterEnabled       bool
	ClusterRoleReadName      string
	ClusterRoleCrudName      string
	RegisteredAdditionalGVKs []schema.GroupVersionKind
	InformersApi             handlers.InformersApi
	CachedAdditionalGvksApi  handlers.CachedAdditionalGvksApi
}

func NewKartaController(client client.Client,
	kartaControlParams *KartaControlParams) *KartaController {
	registeredAdditionalGVKsMap := make(map[schema.GroupVersionKind]bool)
	for _, gvk := range kartaControlParams.RegisteredAdditionalGVKs {
		registeredAdditionalGVKsMap[gvk] = true
	}

	rbacHandler := handlers.NewKartaRBACHandler(client, kartaControlParams.CachedAdditionalGvksApi, kartaControlParams.RBACUpdaterEnabled,
		kartaControlParams.ClusterRoleReadName, kartaControlParams.ClusterRoleCrudName)
	deletionHandler := handlers.NewKartaDeletionHandler(client, rbacHandler, registeredAdditionalGVKsMap, kartaControlParams.InformersApi)

	return &KartaController{
		Client: client,

		rbacHandler:              rbacHandler,
		registeredAdditionalGVKs: registeredAdditionalGVKsMap,
		deletionHandler:          deletionHandler,

		rbacUpdaterEnabled:  kartaControlParams.RBACUpdaterEnabled,
		clusterRoleReadName: kartaControlParams.ClusterRoleReadName,
		clusterRoleCrudName: kartaControlParams.ClusterRoleCrudName,
	}
}

func (rc *KartaController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	karta := &kartav1alpha1.Karta{}
	if shouldContinue, err := rc.getKartaForRequest(ctx, req, karta); err != nil {
		return ctrl.Result{}, err
	} else if !shouldContinue {
		return ctrl.Result{}, nil
	}

	var err error

	if karta.DeletionTimestamp.IsZero() {
		err = rc.addControllerAsFinalizerIfNeeded(ctx, karta)
		if err != nil {
			return ctrl.Result{}, err
		}

		log.Info().Msgf("Reconciling Karta <%s>", karta.Name)
		err = rc.reconcile(ctx, karta)
	} else {
		err = rc.finalize(ctx, karta)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (rc *KartaController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&kartav1alpha1.Karta{})

	if !rc.rbacUpdaterEnabled {
		// since no escalate, we need to watch the 2 ClusterRoles to be aware of manual changes that should be made
		builder = builder.
			Watches(&rbacv1.ClusterRole{}, handler.EnqueueRequestsFromMapFunc(rc.MapClusterRoleToKartaEvent)).
			WithEventFilter(rc.ControllerPredicateFuncs())
	}

	return builder.
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(rc)
}

func (rc *KartaController) getKartaForRequest(ctx context.Context, req ctrl.Request, karta *kartav1alpha1.Karta) (shouldContinue bool, err error) {
	err = rc.Client.Get(ctx, req.NamespacedName, karta)
	if err == nil {
		return true, nil
	}

	if errors.IsNotFound(err) {
		log.Error().Msgf("Can't locate Karta <%s> in cluster, this might also mean it was purposefully deleted", req.Name)
		// Nothing really more to do.
		return false, nil
	}
	log.Error().Msgf("Could not extract Karta <%s> from incoming request: %v, err: %v", req.Name, req, err.Error())
	return false, err
}
