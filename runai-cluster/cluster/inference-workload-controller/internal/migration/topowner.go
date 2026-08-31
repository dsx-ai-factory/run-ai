/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v3"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/spec_checksum"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TopOwnerMigrator is a migration that takes care of the following:
// 1. remove the finalizer set on the inference workload
// 2. add the serving port service type to the workload port, if needed
// 3. add a spec checksum annotation to the inference workload
// 4. set the inference workload as the owner of the knative service
// 5. delete external workload (ew) resources and handle related changes in pod and podgroup metadata
// these actions migrate a workload to be aligned with workloads that are
// handled by the inference workload controller of version 2.19
type TopOwnerMigrator struct {
	k8sClient                 client.Client
	k8sClientWithFieldIndexer CachedClient
}

func NewTopOwnerMigrator(k8sClient client.Client, k8sClientWithFieldIndexer CachedClient) *TopOwnerMigrator {
	return &TopOwnerMigrator{
		k8sClient:                 k8sClient,
		k8sClientWithFieldIndexer: k8sClientWithFieldIndexer,
	}
}

func (tom *TopOwnerMigrator) String() string {
	return TopOwnerMigrationName
}

func (tom *TopOwnerMigrator) RequiresWebhooks() bool {
	return true
}

// Migrate iterates all the runai namespaces, with pagination, and calls migrateWorkloadInNamespace on each of them
func (tom *TopOwnerMigrator) Migrate(ctx context.Context) error {
	namespaceList := &corev1.NamespaceList{}
	var continueToken string
	var migratedWorkloadsCounter uint64

	defer func() {
		log.Info("migration completed", "total workloads migrated", migratedWorkloadsCounter)
	}()

	for {

		if ctx.Err() != nil {
			return fmt.Errorf("context is cancelled; %w", ctx.Err())
		}

		pagination := client.ListOptions{
			Limit:    paginationLimit,
			Continue: continueToken,
		}

		err := tom.k8sClient.List(ctx, namespaceList, &pagination, client.HasLabels{commonconstants.RunaiQueueLabel})
		if err != nil {
			return fmt.Errorf("unable to list namespaces with runai label; %w", err)
		}

		for _, ns := range namespaceList.Items {
			nsCP := ns
			numMigratedWorkloads, err := tom.migrateWorkloadInNamespace(ctx, nsCP.Name)
			if err != nil {
				return err
			}

			migratedWorkloadsCounter += numMigratedWorkloads
		}

		continueToken = namespaceList.Continue
		if continueToken == "" {
			break
		}
	}

	return nil
}

// migrateWorkloadInNamespace iterates all the workloads in a namespace, with pagination, and calls migrateWorkload on
// each of them. it returns the number of workloads migrated and an error if any
func (tom *TopOwnerMigrator) migrateWorkloadInNamespace(ctx context.Context, namespace string) (uint64, error) {
	var migratedWorkloadsCounter uint64
	var continueToken string

	defer func() {
		log.Info("migration completed for namespace", "namespace", namespace, "num workloads migrated", migratedWorkloadsCounter)
	}()

	log.Info("migrating workloads in namespace", "namespace", namespace)

	for {
		if ctx.Err() != nil {
			return migratedWorkloadsCounter, fmt.Errorf("context is cancelled; %w", ctx.Err())
		}

		nextBatch, newContinueToken, err := tom.getNextWorkloadBatch(ctx, namespace, continueToken)
		if err != nil {
			return migratedWorkloadsCounter, err
		}

		for _, workload := range nextBatch {
			log.Info("migrating workload", "workload", workload.Name)
			err := tom.migrateWorkload(ctx, &workload)
			if err != nil {
				return migratedWorkloadsCounter, fmt.Errorf("failed to execute %s migration on workload %s: %w", tom.String(), workload.Name, err)
			}
		}

		migratedWorkloadsCounter += uint64(len(nextBatch))

		continueToken = newContinueToken
		if continueToken == "" {
			break
		}
	}

	return migratedWorkloadsCounter, nil
}

// getNextWorkloadBatch returns the next batch of workloads to migrate by listing all the workloads in the namespace,
// with a continue token (for pagination)
func (tom *TopOwnerMigrator) getNextWorkloadBatch(ctx context.Context, namespace string, continueToken string) ([]infw.InferenceWorkload, string, error) {
	workloads := &infw.InferenceWorkloadList{}

	pagination := client.ListOptions{
		Limit:    paginationLimit,
		Continue: continueToken,
	}

	err := tom.k8sClient.List(ctx, workloads, &pagination, client.InNamespace(namespace))
	if err != nil {
		if meta.IsNoMatchError(err) {
			log.Info("Skipping top-owner migration because the InferenceWorkload API is not served", "namespace", namespace)
			return nil, "", nil
		}
		return nil, "", err
	}

	return workloads.Items, workloads.Continue, nil
}

// migrateWorkload migrates a single workload by:
// 1. removing the finalizer set on the workload
// 2. adding the serving port service type to the workload port, if needed (older versions didn't have it)
// 3. setting the spec checksum annotation on the workload
// 4. setting the workload as the owner of the knative service, if such
// 5. delete external workload (ew) resources and handle related changes in pod and podgroup metadata
// the order of the actions is important, as some of them depend on the state of the workload after the previous actions!
// especially, handling the external workload (and the pod label) must be done after setting the inference workload as
// the owner, otherwise the external-workload-integrator might handle the knative service as an external-workload
func (tom *TopOwnerMigrator) migrateWorkload(ctx context.Context, workload *infw.InferenceWorkload) error {
	err := tom.handleInferenceWorkloadResourceMigrations(ctx, workload)
	if err != nil {
		return err
	}

	err = tom.handleExternalWorkloadResourceMigrations(ctx, workload)
	if err != nil {
		return err
	}

	return nil
}

// handleInferenceWorkloadResourceMigrations takes care of the migration steps that affect the InferenceWorkload resource
func (tom *TopOwnerMigrator) handleInferenceWorkloadResourceMigrations(ctx context.Context, workload *infw.InferenceWorkload) error {
	log.Info("handling inference workload resource migrations", "workload", workload.Name)

	// if spec-checksum is already set, we don't need to do anything
	if spec_checksum.GetSpecChecksumAnnotation(workload) != "" {
		log.Info("workload already has spec checksum annotation set, nothing to do", "workload", workload.Name)
		return nil
	}

	// handle finalizers
	tom.removeFinalizers(workload)

	// handle serving port type
	tom.addServingPortType(workload)

	knativeServiceFound, err := tom.setKnativeServiceOwnerReference(ctx, workload)
	if err != nil {
		return err
	}

	// if no knative service was found, we should not set the spec checksum annotation,
	// as we want the controller to build the missing knative service resource for the workload!
	//
	// important: this stage must be done after all the other workload-related stages, as we want the spec-checksum
	// to be calculated on the final spec of the workload
	if knativeServiceFound {
		_, err := spec_checksum.SetSpecChecksumAnnotation(workload, "")
		if err != nil {
			return err
		}

		_ = spec_checksum.SetSpecChecksumMigratedAnnotation(workload, true)
	}

	// don't apply further changes on the workload at this point! only before calculating the spec-checksum

	// update the workload object
	err = updateWorkloadObject(ctx, tom.k8sClient, workload)
	if err != nil {
		return err
	}

	return nil
}

// setKnativeServiceOwnerReference sets the inference workload as the owner of the knative service, if such exists.
// it returns a boolean indicating if the knative service was found, and an error if any
func (tom *TopOwnerMigrator) setKnativeServiceOwnerReference(ctx context.Context, workload *infw.InferenceWorkload) (bool, error) {
	knativeService, err := tom.getWorkloadKnativeService(workload)
	if err != nil {
		return false, err
	}

	if knativeService == nil {
		log.Info("no knative service found for workload, nothing to do", "workload", workload.Name)
		return false, nil
	}

	log.Info("setting owner for knative service", "workload", workload.Name)
	err = controller_utils.SetControllerReference(workload, knativeService, tom.k8sClient.Scheme())
	if err != nil {
		return true, err
	}

	log.Info("updating knative service", "workload", workload.Name)
	err = tom.k8sClient.Update(ctx, knativeService)
	if err != nil {
		return true, err
	}

	return true, nil
}

func (tom *TopOwnerMigrator) getWorkloadKnativeService(workload *infw.InferenceWorkload) (*knv1.Service, error) {
	knativeService := &knv1.Service{}
	err := tom.k8sClient.Get(context.Background(), client.ObjectKey{
		Namespace: workload.Namespace,
		Name:      workload.Name,
	}, knativeService)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("knative service not found", "workload", workload.Name)
			return nil, nil
		} else {
			log.Error(err, "failed to get knative service", "workload", workload.Name)
			return nil, err
		}
	}

	return knativeService, nil
}

func (tom *TopOwnerMigrator) removeFinalizers(workload *infw.InferenceWorkload) {
	workload.SetFinalizers([]string{})
}

// addServingPortType adds the serving port service type to the workload, in the relevant port.
// it only does that if the service type is not set on the port, and the default service type is not set on the workload
func (tom *TopOwnerMigrator) addServingPortType(workload *infw.InferenceWorkload) {
	defaultServiceTypeExists := workload.Spec.ServiceType.IsValueInitialized()

	for _, port := range workload.Spec.Ports.Deref() {
		if port.Value.ServiceType == nil && !defaultServiceTypeExists {
			port.Value.ServiceType = utils.Ref(commonconstants.ServiceTypeServingPort)
		}
	}
}

// updateWorkloadObject updates the workload object in the cluster, with several attempts.
// in some cases the Update call fails due to webhooks not being ready ("connection refused") and therefore we retry
func updateWorkloadObject(ctx context.Context, k8sClient client.Client, workload *infw.InferenceWorkload) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 50 * time.Millisecond
	expBackoff.MaxInterval = 400 * time.Millisecond
	expBackoff.MaxElapsedTime = 2 * time.Second

	attempt := 0
	ticker := backoff.NewTicker(expBackoff)
	defer ticker.Stop()

	var err error
	for range ticker.C {
		err = k8sClient.Update(ctx, workload)
		if err == nil {
			log.Info("workload updated successfully", "workload", workload.Name)
			return nil
		}

		log.Info("failed to update workload object, retrying...", "workload", workload.Name, "error", err, "attempt", attempt)

		attempt += 1
	}

	log.Info("failed to update workload object after all attempts", "workload", workload.Name, "error", err)

	return err
}

// handleExternalWorkloadResourceMigrations takes care of the migration steps that affect the ExternalWorkload,
// Pod and PodGroup resources
func (tom *TopOwnerMigrator) handleExternalWorkloadResourceMigrations(ctx context.Context, workload *infw.InferenceWorkload) error {
	log.Info("handling external workload resource migrations", "workload", workload.Name)

	// delete the external workload resource
	err := tom.deleteExternalWorkload(ctx, workload)
	if err != nil {
		return err
	}

	// handle pods and podgroup metadata
	err = tom.handlePodsAndPodGroupMetadata(ctx, workload)
	if err != nil {
		return err
	}

	return nil
}

func (tom *TopOwnerMigrator) deleteExternalWorkload(ctx context.Context, workload *infw.InferenceWorkload) error {
	// get the external workload resource. we do it by a field selector, looking for an ew with "workload original uid"
	// that equals to the uid of workload's knative service

	// first, get the knative service
	knativeService, err := tom.getWorkloadKnativeService(workload)
	if err != nil {
		return err
	}

	if knativeService == nil {
		log.Info("no knative service found for workload, nothing to do with ew", "workload", workload.Name)
		return nil
	}

	// now, look for the ew using the field selector with the indexer that was set during the migration setup
	externalWorkloadByOriginalFieldSelector :=
		fields.OneTermEqualSelector(ExternalWorkloadByOriginalUIDIndexerName, string(knativeService.GetUID()))

	externalWorkloads := &v1alpha1.ExternalWorkloadList{}
	err = tom.k8sClientWithFieldIndexer.List(ctx, externalWorkloads,
		&client.ListOptions{FieldSelector: externalWorkloadByOriginalFieldSelector},
		client.InNamespace(workload.Namespace))
	if err != nil {
		log.Error(err, "failed to list external workloads for workload", "workload", workload.Name, "knative service uid", knativeService.GetUID())
		return err
	}

	// finally, delete the external workload resources (expecting 0 or 1)
	for _, ew := range externalWorkloads.Items {
		log.Info("deleting external workload", "workload", workload.Name, "external workload", ew.Name)
		err = tom.k8sClient.Delete(ctx, &ew)
		if err != nil {
			log.Error(err, "failed to delete external workload", "workload", workload.Name, "external workload", ew.Name)
			return err
		}
	}

	return nil
}

func (tom *TopOwnerMigrator) handlePodsAndPodGroupMetadata(ctx context.Context, workload *infw.InferenceWorkload) error {
	// get the pods and podgroups that are related to the workload. the label we use exists (at least) since version 2.13
	listOptionByNamespace := client.InNamespace(workload.Namespace)
	listOptionByLabels := client.MatchingLabels{
		LabelInferenceWorkload: workload.GetChildName(),
	}
	listOptions := []client.ListOption{listOptionByNamespace, listOptionByLabels}

	pods := &corev1.PodList{}
	err := tom.k8sClient.List(ctx, pods, listOptions...)
	if err != nil {
		log.Error(err, "failed to list pods for workload", "workload", workload.Name)
		return err
	}

	podGroups := &kaiv2alpha2.PodGroupList{}
	err = tom.k8sClient.List(ctx, podGroups, listOptions...)
	if err != nil {
		log.Error(err, "failed to list pod groups for workload", "workload", workload)
		return err
	}

	// pods: replace the top-owner-uid label with the workload uid (instead of the knative service one)
	for _, pod := range pods.Items {
		log.Info("patching pod with top-owner-uid label", "pod", pod.Name, "workload", workload.Name)
		err = tom.k8sClient.Patch(ctx, &pod, client.RawPatch(types.MergePatchType, []byte(
			fmt.Sprintf(`{"metadata":{"labels":{"%s":"%s"}}}`, clustercommonconstants.TopOwnerUidLabel, string(workload.GetUID())))))
		if err != nil {
			log.Error(err, "failed to patch pod with top-owner-uid label", "pod", pod.Name, "workload", workload.Name)
			return err
		}
	}

	// podgroups: remove the external-workload-name annotation
	for _, podGroup := range podGroups.Items {
		log.Info("patching podgroup with external-workload-name annotation", "pod group", podGroup.Name, "workload", workload.Name)
		err = tom.k8sClient.Patch(ctx, &podGroup, client.RawPatch(types.MergePatchType, []byte(
			fmt.Sprintf(`{"metadata":{"annotations":{"%s":null}}}`, clustercommonconstants.ExternalWorkloadName))))
		if err != nil {
			log.Error(err, "failed to patch podgroup with external-workload-name annotation", "pod group", podGroup.Name, "workload", workload.Name)
			return err
		}
	}

	return nil
}
