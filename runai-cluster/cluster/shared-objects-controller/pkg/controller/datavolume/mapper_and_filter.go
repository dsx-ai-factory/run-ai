/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume

import (
	"context"
	"reflect"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (c *DatavolumeReconciler) MapPvcToDatavolumeEvent(_ context.Context, object client.Object) []reconcile.Request {
	pvc, ok := object.(*v1.PersistentVolumeClaim)
	if !ok {
		log.Error().Msgf("Expected a PersistentVolumeClaim but got a %T", object)
		return []reconcile.Request{}
	}

	datavolumeName := getDatavolumeNameFromLabels(pvc)
	if datavolumeName != "" {
		log.Debug().Msgf("PVC %s/%s is managed by dataVolume %s", pvc.Namespace, pvc.Name, datavolumeName)
		return []reconcile.Request{{NamespacedName: client.ObjectKey{Name: datavolumeName}}}
	}

	return []reconcile.Request{}
}

func (c *DatavolumeReconciler) MapPvToDatavolumeEvent(_ context.Context, object client.Object) []reconcile.Request {
	pv, ok := object.(*v1.PersistentVolume)
	if !ok {
		log.Error().Msgf("Expected a PersistentVolume but got a %T", object)
		return []reconcile.Request{}
	}

	datavolumeName := getDatavolumeNameFromLabels(pv)
	if datavolumeName != "" {
		log.Debug().Msgf("PV %s is managed by dataVolume %s", pv.Name, datavolumeName)
		return []reconcile.Request{{NamespacedName: client.ObjectKey{Name: datavolumeName}}}
	}

	return []reconcile.Request{}
}

func (c *DatavolumeReconciler) MapProjectToDatavolumeEvent(ctx context.Context, object client.Object) []reconcile.Request {
	project, ok := object.(*kaiorg.Project)
	if !ok {
		log.Error().Msgf("Expected a kai Project but got a %T", object)
		return []reconcile.Request{}
	}

	namespace, err := getters.GetNamespaceByKaiProjectObj(ctx, project, c.client)
	if err != nil {
		log.Error().Msgf("Failed to get namespace by kai project %s, error: %s", project.Name, err.Error())
		return []reconcile.Request{}
	}

	pvcs, err := c.pvcHandler.ListPvcsInNamespace(ctx, namespace.Name)
	if err != nil {
		log.Error().Msgf("Failed to list pvcs in namespace %s, error: %s", namespace.Name, err.Error())
		return []reconcile.Request{}
	}

	pvs, err := c.pvHandler.GetAllDatavolumePvsForNamespace(ctx, namespace.Name)
	if err != nil {
		log.Error().Msgf("Failed to list datavolume pvs for namespace %s, error: %s", namespace.Name, err.Error())
		return []reconcile.Request{}
	}

	dvNamesMap := map[string]bool{}
	setOwnerDatavolumes(pvcs, pvs, dvNamesMap)
	setDatavolumesSharedWithProject(ctx, c.client, project.Name, project.Spec.Parent, dvNamesMap)

	requests := []reconcile.Request{}
	for datavolumeName := range dvNamesMap {
		log.Debug().Msgf("For kai project <%s>, triggering reconcile of datavolume <%s>",
			project.Name, datavolumeName)
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: datavolumeName}})
	}

	return requests
}

func (c *DatavolumeReconciler) DatavolumePredicateFuncs() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return filterCreateEvents(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filterUpdateEvents(e.ObjectOld, e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}
}

func filterCreateEvents(object client.Object) bool {
	switch object.(type) {
	case *v2.Project:
		return object.(*v2.Project).Status.Namespace != ""
	case *kaiorg.Project:
		return object.(*kaiorg.Project).Status.Namespace != ""
	default:
		return true
	}
}

func filterUpdateEvents(objectOld, objectNew client.Object) bool {
	switch objectNew.(type) {
	case *v1alpha1.DataVolume:
		return objectNew.GetResourceVersion() != objectOld.GetResourceVersion()
	case *v1.PersistentVolumeClaim, *v1.PersistentVolume:
		return objectNew.GetResourceVersion() != objectOld.GetResourceVersion() ||
			reflect.DeepEqual(objectNew.GetLabels(), objectOld.GetLabels())
	case *kaiorg.Project:
		return filterProjectUpdatesForController(objectOld.(*kaiorg.Project), objectNew.(*kaiorg.Project))
	default:
		return true
	}
}

func filterProjectUpdatesForController(projectOld, projectNew *kaiorg.Project) bool {
	// we are only interested in Reconcile if the project has a namespace status update
	if projectOld.Status.Namespace != projectNew.Status.Namespace {
		return true
	}
	return false
}

func getDatavolumeNameFromLabels(object client.Object) string {
	name, ok := object.GetLabels()[v1alpha1.DataVolumeNameLabel]
	if ok {
		return name
	}
	return ""
}

func setOwnerDatavolumes(pvcs []v1.PersistentVolumeClaim, pvs []v1.PersistentVolume,
	resultDatavlumeNames map[string]bool) {
	for i := range pvcs {
		pvc := &pvcs[i]
		datavolumeName, ok := pvc.Labels[v1alpha1.DataVolumeNameLabel]
		if ok && datavolumeName != "" {
			resultDatavlumeNames[datavolumeName] = true
		}
	}

	for i := range pvs {
		pv := &pvs[i]
		datavolumeName, ok := pv.Labels[v1alpha1.DataVolumeNameLabel]
		if ok && datavolumeName != "" {
			resultDatavlumeNames[datavolumeName] = true
		}
	}
}

func setDatavolumesSharedWithProject(ctx context.Context, kubeclient client.Client, projectName, departmentName string,
	resultDatavlumeNames map[string]bool) {
	datavolumes, err := listDatavolumes(ctx, kubeclient)
	if err != nil {
		return
	}

	// this is not very efficient but it's the best we can do, an index won't help as well
	for _, datavolume := range datavolumes {
		for _, sharedScope := range datavolume.Spec.SharedScopes {
			if (sharedScope.ScopeType == v1alpha1.ProjectScope && sharedScope.ScopeName == projectName) ||
				(sharedScope.ScopeType == v1alpha1.DepartmentScope && sharedScope.ScopeName == departmentName) ||
				sharedScope.ScopeType == v1alpha1.ClusterScope {
				resultDatavlumeNames[datavolume.Name] = true
				break
			}
		}
	}
}

func listDatavolumes(ctx context.Context, kubeclient client.Client) ([]v1alpha1.DataVolume, error) {
	datavolumes := v1alpha1.DataVolumeList{}
	err := kubeclient.List(ctx, &datavolumes)
	if err != nil {
		log.Error().Msgf("Failed to list datavolumes, error: %s", err.Error())
		return nil, err
	}
	return datavolumes.Items, nil
}
