/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"strconv"

	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	constants2 "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddRevisionHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) *cachedclient.RateLimitedController {
	revisionInformer := cachedClient.GetInformer(cachedclient.RevisionType)
	if revisionInformer == nil {
		// assumes knative is not installed
		return nil
	}

	revisionHandler := handlers.NewRevisionHandler(cachedClient, backendClient)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.RevisionRL, revisionHandler)
	backendClient.RegisterSyncHandler(revisionHandler)

	_, err := revisionInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(object interface{}) bool {
			var rev *knv1.Revision
			var ok bool
			if rev, ok = object.(*knv1.Revision); !ok || rev == nil {
				return false
			}

			return commonconstants.IsManagedInferenceRevision(rev.Labels)
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(object interface{}) {
				obj := object.(*knv1.Revision)
				genericControllerErrorHandler(revisionHandler.HandleRevisionAdd(obj), constants2.AddOperation, constants2.RevisionResourceName, obj)
			},
			DeleteFunc: func(object interface{}) {
				obj := object.(*knv1.Revision)
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
				genericControllerErrorHandler(revisionHandler.HandleRevisionDelete(object.(client.Object)), constants2.DeleteOperation, constants2.RevisionResourceName, obj)
			},
			UpdateFunc: func(oldObject interface{}, newObject interface{}) {
				oldRevision := oldObject.(*knv1.Revision)
				newRevision := newObject.(*knv1.Revision)

				// Only proceed if there are status changes we care about
				oldPhaseVersion := getPhaseVersion(oldRevision.Status.Annotations, "old")
				newPhaseVersion := getPhaseVersion(newRevision.Status.Annotations, "new")

				if oldPhaseVersion >= newPhaseVersion {
					return
				}

				genericControllerErrorHandler(revisionHandler.HandleRevisionUpdate(newRevision), constants2.UpdateOperation, constants2.RevisionResourceName, newRevision)
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to revision informer, error: %s", err.Error())
	}

	zap.S().Debug("The revision controller has started")
	return nil
}

func getPhaseVersion(revisionStatusAnnotations map[string]string, oldNewStr string) int {
	phaseVersionStr, ok := revisionStatusAnnotations[constants.RunaiPhaseVersionAnnotationKey]
	if !ok || phaseVersionStr == "" {
		zap.S().Debugf("Failed to get phase version from %s revision status annotations, assuming 0", oldNewStr)
		return 0
	}

	phaseVersion, err := strconv.Atoi(phaseVersionStr)
	if err != nil {
		zap.S().Warnf("Failed to convert %s phase version '%s' to int, assuming 0, error: %s", oldNewStr, phaseVersionStr, err.Error())
		return 0
	}
	return phaseVersion
}
