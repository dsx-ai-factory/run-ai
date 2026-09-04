/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/shared-objects-controller/pkg/controller/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/shared-objects-controller/pkg/controller/datavolume/handlers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	randomSuffixLen                    = 4
	MaxNameLenForPvc                   = 50 // with buffer for "-" and suffix
	MaxNameLenForPv                    = 25 // with buffer for "-" and suffix
	ClusterAndDepartmentScopeNamespace = "runai"
)

type DatavolumeReconciler struct {
	client client.Client
	common.CommonReconciler

	pvcHandler *handlers.PvcHandler
	pvHandler  *handlers.PvHandler
}

func NewDatavolumeReconciler(client client.Client) *DatavolumeReconciler {
	datavolumeReconciler := &DatavolumeReconciler{
		client:     client,
		pvcHandler: handlers.NewPvcHandler(client),
		pvHandler:  handlers.NewPvHandler(client),
	}
	datavolumeReconciler.CommonReconciler = common.NewCommonReconciler(client, FinalizerName, datavolumeReconciler)

	return datavolumeReconciler
}

func (c *DatavolumeReconciler) ReconcileObject(ctx context.Context, object client.Object) (client.Object, error) {
	datavolume, ok := object.(*v1alpha1.DataVolume)
	if !ok {
		log.Ctx(ctx).Error().Msgf("Expected DataVolume but got: %T", object)
		return object, nil
	}

	datavolumeWithUpdatedStatus, err := c.reconcileDatavolume(ctx, datavolume)
	if err != nil {
		datavolumeWithUpdatedStatus.Status.Phase = v1alpha1.DataVolumeNotReady
	} else {
		datavolumeWithUpdatedStatus.Status.Phase = v1alpha1.DataVolumeReady
		datavolumeWithUpdatedStatus.Status.Message = ""
	}

	return datavolumeWithUpdatedStatus, err
}

func (c *DatavolumeReconciler) reconcileDatavolume(ctx context.Context, datavolume *v1alpha1.DataVolume) (*v1alpha1.DataVolume, error) {
	datavolumeWithUpdatedStatus := datavolume.DeepCopy()
	datavolumeObjectsToReconcile := common.InitEmptyObjectsToReconcile()

	var err error

	datavolumeStatus, setObjectsErr := c.setObjectsToReconcile(ctx, datavolume, datavolumeObjectsToReconcile)
	datavolumeWithUpdatedStatus.Status = *datavolumeStatus
	if setObjectsErr != nil {
		err = multierror.Append(err, setObjectsErr)
	}

	// continue even if failed - perhaps we have some objects to clean up
	statusMsg, innerErr := c.reconcilePvcsAndPvs(ctx, datavolumeObjectsToReconcile)
	if innerErr != nil {
		statusMessages := make([]string, 0, 2)
		if setObjectsErr != nil && datavolumeStatus.Message != "" {
			statusMessages = append(statusMessages, datavolumeStatus.Message)
		}
		if statusMsg != "" {
			statusMessages = append(statusMessages, string(statusMsg))
		}
		datavolumeWithUpdatedStatus.Status.Message = strings.Join(statusMessages, ";")
		err = multierror.Append(err, innerErr)
	}

	return datavolumeWithUpdatedStatus, err
}

func (c *DatavolumeReconciler) setObjectsToReconcile(ctx context.Context, datavolume *v1alpha1.DataVolume, objectsToReconcile *common.ObjectsToReconcile) (*v1alpha1.DataVolumeStatus, error) {
	status := datavolume.Status.DeepCopy()

	originalNamespace, errStr, err := c.getNamespaceOfOriginalPvc(ctx, datavolume)
	if err != nil {
		status.Message = string(errStr)
		return status, err
	}

	datavolumePvc, datavolumePv, errStr, err := c.getAndValidateDatavolumePvcAndPv(ctx, datavolume, originalNamespace, objectsToReconcile)
	if err != nil {
		status.Message = string(errStr)
		return status, err
	}

	status.DatavolumePvcName = datavolumePvc.Name
	status.DatavolumePvName = datavolumePv.Name

	namespacesForScopes, err := c.getNamespacesForScopes(ctx, datavolume)
	if err != nil {
		if errors.IsNotFound(err) {
			status.Message = string(ScopeNotFound)
		} else {
			status.Message = string(InternalError)
		}
		return status, err
	}

	err = c.setDatavolumeObjectsToReconcile(ctx, datavolume, datavolumePvc, datavolumePv, namespacesForScopes, objectsToReconcile)
	if err != nil {
		status.Message = string(InternalError)
		return status, err
	}

	return status, nil
}

func (c *DatavolumeReconciler) reconcilePvcsAndPvs(ctx context.Context, datavolumeObjectsToReconcile *common.ObjectsToReconcile) (DatavolumePhaseMessage, error) {
	var result error
	var statusMessage DatavolumePhaseMessage

	if err := c.addPvcsBeforePvs(ctx, datavolumeObjectsToReconcile); err != nil {
		result = multierror.Append(result, err)
		statusMessage = FailedToCreatePvcsPvs
	}

	if err := c.UpdateObjects(ctx, datavolumeObjectsToReconcile); err != nil {
		result = multierror.Append(result, err)
		if statusMessage == "" {
			statusMessage = FailedToUpdatePvcsPvs
		}
	}

	if err := c.DeleteObjects(ctx, datavolumeObjectsToReconcile); err != nil {
		result = multierror.Append(result, err)
		if statusMessage == "" {
			statusMessage = FailedToDeletePvcsPvs
		}
	}

	return statusMessage, result
}

func (c *DatavolumeReconciler) addPvcsBeforePvs(ctx context.Context, objectsToReconcile *common.ObjectsToReconcile) error {
	var result error
	failedPvcs := make(map[client.ObjectKey]bool)

	for _, object := range objectsToReconcile.ToAdd {
		switch typedObject := object.(type) {
		case *v1.PersistentVolumeClaim:
			if err := c.Create(ctx, typedObject); err != nil {
				failedPvcs[client.ObjectKeyFromObject(typedObject)] = true
				result = multierror.Append(result, err)
			}
		case *v1.PersistentVolume:
			if typedObject.Spec.ClaimRef != nil && failedPvcs[client.ObjectKey{
				Name:      typedObject.Spec.ClaimRef.Name,
				Namespace: typedObject.Spec.ClaimRef.Namespace,
			}] {
				log.Ctx(ctx).Warn().Msgf(
					"Skipping creation of pv %s because pvc %s/%s failed to be created",
					typedObject.Name, typedObject.Spec.ClaimRef.Namespace, typedObject.Spec.ClaimRef.Name)
				continue
			}
			if err := c.Create(ctx, typedObject); err != nil {
				result = multierror.Append(result, err)
			}
		default:
			if err := c.Create(ctx, object); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result
}

// getAndValidateDatavolumePvcAndPv -
// if the datavolume's pvc and pv COPIES already exist in the original namespace - return them;
// -> because we allow the original ones to be deleted;
// if they don't exist - validate the original user-created pvc and pv; and construct the copies.
// update the "datavolumeObjectsToReconcile" with the pvc and pv - toAdd if needed.
func (c *DatavolumeReconciler) getAndValidateDatavolumePvcAndPv(ctx context.Context, datavolume *v1alpha1.DataVolume,
	originalNamespace string, datavolumeObjectsToReconcile *common.ObjectsToReconcile) (*v1.PersistentVolumeClaim, *v1.PersistentVolume, DatavolumePhaseMessage, error) {
	var datavolumePvc *v1.PersistentVolumeClaim
	var datavolumePv *v1.PersistentVolume
	var err error

	datavolumePvcName := datavolume.Status.DatavolumePvcName
	if datavolumePvcName != "" {
		// first, try to get the already-copied pvc and pv in the original namespace
		datavolumePvc, datavolumePv, err = c.getPvcAndItsPv(ctx, datavolumePvcName, originalNamespace)
		if err != nil {
			return nil, nil, InternalError, err
		}
		if datavolumePvc != nil && datavolumePv != nil {
			if datavolume.IsOwner(datavolumePvc) == 0 && datavolume.IsOwner(datavolumePv) == 0 {
				return datavolumePvc, datavolumePv, "", nil
			}
		}
	}

	// validate if they already exist but were not reflected in status
	datavolumePvc, datavolumePv, err = c.findDatavolumeCopiedOriginalPvPvc(ctx, datavolume, originalNamespace, datavolumeObjectsToReconcile)
	if err != nil {
		return nil, nil, InternalError, err
	}
	if datavolumePvc != nil && datavolumePv != nil {
		return datavolumePvc, datavolumePv, "", nil
	}
	if datavolumePvc != nil && datavolumePv == nil {
		// shouldn't happen, but if it does - remove it
		datavolumeObjectsToReconcile.ToRemove = append(datavolumeObjectsToReconcile.ToRemove, datavolumePvc)
	}
	if datavolumePvc == nil && datavolumePv != nil {
		// shouldn't happen, but if it does - remove it
		datavolumeObjectsToReconcile.ToRemove = append(datavolumeObjectsToReconcile.ToRemove, datavolumePv)
	}

	// if we got here, we need to either use the original PVC/PV or create new copies
	originalPvc, originalPv, err := c.getPvcAndItsPv(ctx, datavolume.Spec.OriginPvcName, originalNamespace)
	if err != nil {
		return nil, nil, InternalError, err
	}

	if originalPvc == nil {
		return nil, nil, OriginalPvcDoesntExist, fmt.Errorf("original pvc %s/%s not found", originalNamespace, datavolume.Spec.OriginPvcName)
	}
	if originalPvc.Status.Phase != v1.ClaimBound {
		return nil, nil, OriginalPvcNotBound, fmt.Errorf("original pvc %s/%s is not bound to a pv", originalNamespace, datavolume.Spec.OriginPvcName)
	}
	if originalPv == nil {
		return nil, nil, PvOfOriginalPvcDoesntExist, fmt.Errorf("original pv %s not found", originalPvc.Spec.VolumeName)
	}
	if originalPv.Status.Phase != v1.VolumeBound {
		return nil, nil, PvOfOriginalPvcNotBound, fmt.Errorf("original pv %s is not bound to a pvc", originalPv.Name)
	}

	// If shouldUseOriginalVolume is true, add labels to the original PVC and PV and return them
	if datavolume.Spec.ShouldUseOriginalVolume {
		// Add labels to original PVC
		if originalPvc.Labels == nil {
			originalPvc.Labels = make(map[string]string)
		}
		originalPvc.Labels[v1alpha1.DataVolumeIdLabel] = datavolume.Spec.DatavolumeId
		originalPvc.Labels[v1alpha1.DataVolumeNameLabel] = datavolume.Name
		datavolumeObjectsToReconcile.ToUpdate = append(datavolumeObjectsToReconcile.ToUpdate, originalPvc)

		// Add labels to original PV
		if originalPv.Labels == nil {
			originalPv.Labels = make(map[string]string)
		}
		originalPv.Labels[v1alpha1.DataVolumeIdLabel] = datavolume.Spec.DatavolumeId
		originalPv.Labels[v1alpha1.DataVolumeNameLabel] = datavolume.Name
		originalPv.Labels[v1alpha1.DataVolumePvForNamespaceLabel] = originalNamespace
		datavolumeObjectsToReconcile.ToUpdate = append(datavolumeObjectsToReconcile.ToUpdate, originalPv)

		return originalPvc, originalPv, "", nil
	}

	// Otherwise, create new copies
	generatedPvcName := generatePvcName(datavolume, &datavolumePvcName)
	generatedPvName := generatePvName(datavolume, originalNamespace)

	datavolumeNewPvc := c.pvcHandler.ConstructCopiedPvcInNamespace(originalPvc, datavolume, generatedPvcName, originalNamespace, generatedPvName)
	datavolumeNewPv := c.pvHandler.ConstructCopiedPv(originalPv, datavolume, generatedPvName, generatedPvcName, originalNamespace, true)
	datavolumeObjectsToReconcile.ToAdd = append(datavolumeObjectsToReconcile.ToAdd, datavolumeNewPvc, datavolumeNewPv)

	return datavolumeNewPvc, datavolumeNewPv, "", nil
}

// findDatavolumeCopiedOriginalPvPvc - finds the copied pvc and pv of the datavolume in the original namespace.
// if only one of them is found, or they are not bound - they will be removed.
// so the function will either return both ov+pvc, or none.
func (c *DatavolumeReconciler) findDatavolumeCopiedOriginalPvPvc(ctx context.Context, datavolume *v1alpha1.DataVolume,
	originalNamespace string, datavolumeObjectsToReconcile *common.ObjectsToReconcile) (
	*v1.PersistentVolumeClaim, *v1.PersistentVolume, error) {
	pvcFound := true

	pvc, err := c.getDatavolumePvcInNamespace(ctx, datavolume, originalNamespace, datavolumeObjectsToReconcile)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, nil, err
		}
		pvcFound = false
	}

	// maybe the pvc doesn't exist but the pv does - then we should find it and delete it

	pv, err := c.getDatavolumePvForNamespace(ctx, datavolume, originalNamespace, datavolumeObjectsToReconcile)
	if err != nil {
		if pvcFound {
			// remove the pvc if the pv doesn't exist - will create a new one
			datavolumeObjectsToReconcile.ToRemove = append(datavolumeObjectsToReconcile.ToRemove, pvc)
			log.Warn().Msgf("Datavolume %s has copied pvc for namespace %s that doesn't have a pv - removing it",
				datavolume.Name, originalNamespace)
		}

		if errors.IsNotFound(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	if !pvcFound {
		// remove the pv if the pvc doesn't exist - will create a new one
		datavolumeObjectsToReconcile.ToRemove = append(datavolumeObjectsToReconcile.ToRemove, pv)
		log.Warn().Msgf("Datavolume %s has copied pv for namespace %s that doesn't have a pvc - removing it",
			datavolume.Name, originalNamespace)
		return nil, nil, nil
	}

	if pv.Spec.ClaimRef == nil ||
		pv.Spec.ClaimRef.Name != pvc.Name || pv.Spec.ClaimRef.Namespace != pvc.Namespace ||
		pvc.Spec.VolumeName != pv.Name {
		// if the pv and pvc are not bound - remove them
		datavolumeObjectsToReconcile.ToRemove = append(datavolumeObjectsToReconcile.ToRemove, pvc, pv)
		log.Warn().Msgf("Datavolume %s has copied pvc and pv for namespace %s that are not bound - removing them",
			datavolume.Name, originalNamespace)
		return nil, nil, nil
	}

	return pvc, pv, nil
}

// getPvcAndItsPv - gets the pvc and its pv.
// notice both return values can be nil if not found, even if error is nil.
func (c *DatavolumeReconciler) getPvcAndItsPv(ctx context.Context, pvcName, namespace string) (
	*v1.PersistentVolumeClaim, *v1.PersistentVolume, error) {
	pvc, err := c.pvcHandler.GetPvc(ctx, pvcName, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	pv, err := c.pvHandler.GetPvOfPvc(ctx, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			return pvc, nil, nil
		}
		return nil, nil, err
	}

	return pvc, pv, nil
}

// setDatavolumeObjectsToReconcile -
// will calculate all needed objects to be created/removed for the datavolume's pvc and pv;
// except the ones in the original namespace - they should have been handled already.
func (c *DatavolumeReconciler) setDatavolumeObjectsToReconcile(ctx context.Context,
	datavolume *v1alpha1.DataVolume,
	datavolumePvc *v1.PersistentVolumeClaim, datavolumePv *v1.PersistentVolume,
	namespacesForScopes map[string]bool, result *common.ObjectsToReconcile) error {
	pvNameToPvc, unClaimedPvcs, err := c.pvcHandler.GetPvcsOfDatavolume(ctx, datavolume, ptr.To(datavolumePvc.Namespace))
	if err != nil {
		return err
	}
	pvNameToPv, err := c.pvHandler.GetPvsOfDatavolume(ctx, datavolume, ptr.To(datavolumePvc.Namespace))
	if err != nil {
		return err
	}

	handledNamespaces := map[string]bool{}
	handledPvs := map[string]bool{}

	// first, go over the existing pvc <-> pv pairs
	for pvName, pvc := range pvNameToPvc {
		pv, pvFound := pvNameToPv[pvName]

		if pvFound {
			_, isPvOwnedByDv := pvNameToPv[pv.Name]
			if !isPvOwnedByDv {
				// the pv is not owned by the datavolume - mark not found so we won't remove it
				pvFound = false
				pv = nil
			}
		}

		_, shouldExistInNs := namespacesForScopes[pvc.Namespace]

		if !shouldExistInNs {
			// the pvc is in a namespace where it shouldn't be - remove both pv and pvc
			log.Warn().Msgf("Datavolume %s has pvc %s in namespace %s that shouldn't be there - removing",
				datavolume.Name, pvc.Name, pvc.Namespace)
			result.ToRemove = append(result.ToRemove, pvc)
			if pvFound {
				handledPvs[pvName] = true
				result.ToRemove = append(result.ToRemove, pv)
			}
			continue
		}

		if !pvFound {
			// the pv of this pvc doesn't exist - remove the pvc
			log.Warn().Msgf("Datavolume %s has pvc %s in namespace %s that doesn't have a pv - removing the pvc",
				datavolume.Name, pvc.Name, pvc.Namespace)
			result.ToRemove = append(result.ToRemove, pvc)
			continue
		}

		handledPvs[pvName] = true

		if _, ok := handledNamespaces[pvc.Namespace]; ok {
			// the namespace should already have a pv and pvc - remove the extra ones
			log.Warn().Msgf("Datavolume %s has an extra pvc %s and pv %s in namespace %s - removing both",
				datavolume.Name, pvc.Name, pv.Name, pvc.Namespace)
			result.ToRemove = append(result.ToRemove, pvc, pv)
			continue
		}

		if pv.Spec.ClaimRef != nil &&
			pv.Spec.ClaimRef.Name == pvc.Name &&
			pv.Spec.ClaimRef.Namespace == pvc.Namespace {
			// pv and pvc are bound, keep both
			handledNamespaces[pvc.Namespace] = true

			continue
		}

		// if the pv's ClaimRef does not match the pvc, remove both
		// don't mark the namespace as handled - we need to create a new pvc and pv for it
		log.Warn().Msgf("Datavolume %s has pvc %s in namespace %s that doesn't match its pv %s - removing both",
			datavolume.Name, pvc.Name, pvc.Namespace, pv.Name)
		result.ToRemove = append(result.ToRemove, pvc, pv)
	}

	// the remaining pvs are unclaimed (were not added to handledPvs) - remove them
	for _, pv := range pvNameToPv {
		if _, ok := handledPvs[pv.Name]; !ok {
			log.Warn().Msgf("Datavolume %s has unclaimed pv %s - removing", datavolume.Name, pv.Name)
			result.ToRemove = append(result.ToRemove, pv)
		}
	}

	// now, go over the unclaimed pvcs to remove them
	for _, pvc := range unClaimedPvcs {
		log.Warn().Msgf("Datavolume %s has unclaimed pvc %s in namespace %s - removing", datavolume.Name, pvc.Name, pvc.Namespace)
		result.ToRemove = append(result.ToRemove, pvc)
	}

	// now, go over the namespaces where the pvc should be
	for namespace := range namespacesForScopes {
		_, ok := handledNamespaces[namespace]
		if ok || namespace == datavolumePvc.Namespace {
			// the namespace is already handled, or it's the original namespace
			continue
		}

		// construct a new pvc and pv for the namespace
		generatedPvcName := generatePvcName(datavolume, &datavolumePvc.Name)
		generatedPvName := generatePvName(datavolume, namespace)
		pvc := c.pvcHandler.ConstructCopiedPvcInNamespace(datavolumePvc, datavolume, generatedPvcName, namespace, generatedPvName)
		pv := c.pvHandler.ConstructCopiedPv(datavolumePv, datavolume, generatedPvName, generatedPvcName, namespace, false)

		result.ToAdd = append(result.ToAdd, pvc, pv)
	}

	return nil
}

func (c *DatavolumeReconciler) getNamespaceOfOriginalPvc(ctx context.Context, datavolume *v1alpha1.DataVolume) (string, DatavolumePhaseMessage, error) {
	// If projectName was not provided we assume that the DV is department/cluster scope
	if datavolume.Spec.ProjectName == "" {
		return ClusterAndDepartmentScopeNamespace, "", nil
	}
	project := &kaiorg.Project{}
	err := c.client.Get(ctx, client.ObjectKey{Name: datavolume.Spec.ProjectName}, project)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", ProjectOfOriginalPvcNotFound, err
		}
		return "", InternalError, err
	}

	originalNamespace, err := getters.GetNamespaceByKaiProjectObj(ctx, project, c.client)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", NamespaceOfOriginalPvcNotFound, err
		}
		return "", InternalError, err
	}
	return originalNamespace.Name, "", nil
}

func (c *DatavolumeReconciler) getNamespacesForScopes(ctx context.Context, datavolume *v1alpha1.DataVolume) (map[string]bool, error) {
	return getters.GetNamespacesForKaiScopes(ctx, c.client, datavolume.Spec.SharedScopes, KaiProjectsParentFieldIndexName)
}

// getDatavolumePvcInNamespace - returns the pvc of the datavolume in the namespace.
// if found more than 1, will add them to be removed, and return NotFound error.
func (c *DatavolumeReconciler) getDatavolumePvcInNamespace(ctx context.Context, datavolume *v1alpha1.DataVolume, namespace string,
	datavolumeObjectsToReconcile *common.ObjectsToReconcile) (*v1.PersistentVolumeClaim, error) {
	pvcs, err := c.pvcHandler.ListPvcsOfDatavolumeInNamespace(ctx, namespace, datavolume.Spec.DatavolumeId)
	if err != nil {
		return nil, err
	}

	if len(pvcs) == 1 {
		return &pvcs[0], nil
	}

	if len(pvcs) == 0 {
		return nil, errors.NewNotFound(v1.Resource("persistentvolumeclaim"),
			"no pvc found to have the datavolume label in the namespace")
	}

	for _, pvc := range pvcs {
		datavolumeObjectsToReconcile.ToRemove = append(datavolumeObjectsToReconcile.ToRemove, &pvc)
	}

	log.Warn().Msgf("Datavolume %s has %d pvcs in namespace %s - they will be removed",
		datavolume.Name, len(pvcs), namespace)
	return nil, errors.NewNotFound(v1.Resource("persistentvolumeclaim"),
		"multiple pvcs found to have the datavolume label in the namespace")
}

// getDatavolumePvForNamespace - returns the pv of the datavolume for the namespace.
// if found more than 1, will add them to be removed, and return NotFound error.
func (c *DatavolumeReconciler) getDatavolumePvForNamespace(ctx context.Context, datavolume *v1alpha1.DataVolume, namespace string,
	datavolumeObjectsToReconcile *common.ObjectsToReconcile) (*v1.PersistentVolume, error) {
	pvs, err := c.pvHandler.ListPvsOfDatavolumeForNamespace(ctx, namespace, datavolume.Spec.DatavolumeId)
	if err != nil {
		return nil, err
	}

	if len(pvs) == 1 {
		return &pvs[0], nil
	}

	if len(pvs) == 0 {
		return nil, errors.NewNotFound(v1.Resource("persistentvolume"),
			"no pv found to have the datavolume label for the namespace")
	}

	for _, pv := range pvs {
		datavolumeObjectsToReconcile.ToRemove = append(datavolumeObjectsToReconcile.ToRemove, &pv)
	}

	log.Warn().Msgf("Datavolume %s has %d pvs for namespace %s - they will be removed",
		datavolume.Name, len(pvs), namespace)
	return nil, errors.NewNotFound(v1.Resource("persistentvolume"),
		"multiple pvs found to have the datavolume label for the namespace")
}

func generatePvcName(datavolume *v1alpha1.DataVolume, datavolumePvcName *string) string {
	if datavolumePvcName != nil && *datavolumePvcName != "" {
		return *datavolumePvcName
	}

	dvName := datavolume.Name
	if len(dvName) > MaxNameLenForPvc {
		dvName = dvName[:MaxNameLenForPvc]
	}

	return fmt.Sprintf("dv-%s-%s", dvName, rand.String(randomSuffixLen))
}

func generatePvName(datavolume *v1alpha1.DataVolume, namespace string) string {
	ns := namespace
	if len(ns) > MaxNameLenForPv {
		ns = ns[:MaxNameLenForPv]
	}

	dvName := datavolume.Name
	if len(dvName) > MaxNameLenForPv {
		dvName = dvName[:MaxNameLenForPv]
	}

	return fmt.Sprintf("dv-%s-%s-%s", dvName, ns, rand.String(randomSuffixLen))
}
