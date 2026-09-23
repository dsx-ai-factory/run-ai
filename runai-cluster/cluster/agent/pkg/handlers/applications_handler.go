/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	workloads_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/metrics"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"github.com/hashicorp/go-multierror"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	applicationDesiredPhaseDeleted     = "deleted"
	annotationApplicationUpdateCounter = "run.ai/application-update-counter"
	userCredentialPlaceholderPrefix    = "USERCREDENTIAL:" //nolint:gosec // Protocol marker, not a credential.
	helmOCIURLPrefix                   = "oci://"
)

var errNoMatchingSecret = errors.New("no matching secret found")

type applicationProjectInfo struct {
	namespace      string
	projectName    string
	projectId      string
	departmentId   string
	departmentName string
}

type ApplicationSyncReportInfo struct {
	numApplications int
	theReport       workloads_openapi.ClusterSyncReport
}

type ApplicationsHandler struct {
	*CommonHandler
	syncCounter int64
	syncReport  *ApplicationSyncReportInfo
}

func (h *ApplicationsHandler) MetricLabel() string {
	return "application"
}

func (h *ApplicationsHandler) projectInfoMap() (map[string]applicationProjectInfo, error) {
	projectList := &kaiorg.ProjectList{}
	if err := h.cachedClient.List(context.Background(), projectList); err != nil {
		return nil, err
	}

	result := make(map[string]applicationProjectInfo, len(projectList.Items))
	for _, project := range projectList.Items {
		if project.Status.Namespace == "" {
			continue
		}
		projectId, found := project.Labels[runaiv2.ProjectIdLabel]
		if !found {
			log.Warn().Msgf("Project <%s> does not have a project ID label", project.Name)
			continue
		}
		result[projectId] = applicationProjectInfo{
			namespace:      project.Status.Namespace,
			projectName:    project.Name,
			projectId:      projectId,
			departmentId:   project.Labels[schedv2alpha2.DepartmentIdLabel],
			departmentName: project.Spec.Parent,
		}
	}
	return result, nil
}

func (h *ApplicationsHandler) Handle() error {
	var err error

	if h.syncReport == nil {
		h.syncReport = &ApplicationSyncReportInfo{}
	}

	nextUpdate := h.syncCounter
	var retryAfter *int64
	for {
		backendApplications, innerErr := h.getBackendApplications(nextUpdate)
		if innerErr != nil {
			log.Error().Msgf("Had an error while getting backend applications, sync counter: %d, error: %s", h.syncCounter, innerErr.Error())
			return innerErr
		}

		if backendApplications == nil {
			log.Error().Msgf("Applications were not received from the backend for sync counter %d", h.syncCounter)
			return nil
		}

		var highestUpdate, highestSysUpdate int64
		if backendApplications.ResponseInfo != nil && backendApplications.ResponseInfo.HighestUpdate != nil {
			highestUpdate = *backendApplications.ResponseInfo.HighestUpdate
		}

		if backendApplications.SystemInfo != nil && backendApplications.SystemInfo.HighestUpdate != nil {
			highestSysUpdate = *backendApplications.SystemInfo.HighestUpdate
		}

		numApplications := len(backendApplications.Applications)
		log.Debug().Msgf("Backend applications response - highest update: %d, highest system update: %d, size: %d", highestUpdate, highestSysUpdate, numApplications)

		h.syncReport.numApplications = numApplications

		if numApplications == 0 {
			log.Debug().Msgf("No applications to handle while counter is %d", h.syncCounter)
			break
		}

		h.syncReport.theReport.Updates = make([]workloads_openapi.ClusterSyncReportUpdate, numApplications)

		innerErr = h.processApplications(context.Background(), backendApplications)
		checkpoint := highestUpdate
		if innerErr != nil {
			log.Error().Msgf("Had an error while processing applications, error: %s", innerErr.Error())
			err = multierror.Append(err, innerErr)
			checkpoint = h.syncCounter
		}

		// Retry failed conversions (for example, Secrets still synchronizing) after restart too.
		// Keep fetching later updates while credentials are unavailable.
		failedDeletion := false
		for i, update := range h.syncReport.theReport.Updates {
			if message := update.Message.Get(); message != nil && *message != "" {
				checkpoint = min(checkpoint, max(0, ptr.Deref(update.UpdateCount, h.syncCounter+1)-1))
			}
			if backendApplications.Applications[i].Application.GetDesiredPhase() == applicationDesiredPhaseDeleted && (innerErr != nil || update.Message.Get() != nil) {
				failedDeletion = true
			}
		}
		if retryAfter != nil {
			checkpoint = min(checkpoint, *retryAfter)
		}
		if checkpoint < highestUpdate {
			retryAfter = ptr.To(checkpoint)
		}
		nextUpdate = highestUpdate
		innerErr = h.setSyncCounter(checkpoint)
		if innerErr != nil {
			log.Error().Msgf("Had an error while updating applications sync counter, error: %s", innerErr.Error())
			err = multierror.Append(err, innerErr)
		}

		// The backend also uses updatesAfter to prune acknowledged deletes.
		if failedDeletion {
			break
		}

		if highestUpdate == highestSysUpdate {
			log.Debug().Msgf("No more applications to handle")
			break
		}
	}

	if err != nil {
		log.Error().Msgf("Failed to handle some applications (create or update), error: %s", err.Error())
		return err
	}

	log.Debug().Msgf("Finished handling applications")
	return nil
}

func (h *ApplicationsHandler) processApplications(ctx context.Context, response *workloads_openapi.AIApplicationSyncResponse) error {
	var err error

	projInfoMap, mapErr := h.projectInfoMap()
	if mapErr != nil {
		return mapErr
	}

	clusterUpdates := clusterWorkloadsAndPoliciesUpdates{
		UpdateRef: make([]int, 0, len(response.Applications)),
		ToUpdate:  make([]client.Object, 0, len(response.Applications)),
		DeleteRef: make([]int, 0, len(response.Applications)),
		ToDelete:  make([]types.NamespacedName, 0),
	}

	allObjectsErrors := make([]error, len(response.Applications))

	for idx, appSyncInfo := range response.Applications {
		item := appSyncInfo.Application
		if item == nil {
			log.Error().Msgf("Application sync info at index %d has nil application, skipping", idx)
			allObjectsErrors[idx] = fmt.Errorf("nil application in sync info")
			continue
		}

		projInfo, found := projInfoMap[item.GetProjectId()]
		if !found {
			log.Error().Msgf("Unknown project %s for application %s, skipped", item.GetProjectId(), item.GetName())
			allObjectsErrors[idx] = fmt.Errorf("unknown project %s", item.GetProjectId())
			continue
		}

		if item.GetDesiredPhase() == applicationDesiredPhaseDeleted {
			clusterUpdates.ToDelete = append(clusterUpdates.ToDelete, types.NamespacedName{
				Namespace: projInfo.namespace,
				Name:      item.GetName(),
			})
			clusterUpdates.DeleteRef = append(clusterUpdates.DeleteRef, idx)
			continue
		}

		application, convertErr := h.convertToApplication(ctx, projInfo, item, appSyncInfo.UpdateCount)
		if convertErr != nil {
			log.Error().Msgf("Failed to convert application %s: %s", item.GetName(), convertErr.Error())
			allObjectsErrors[idx] = convertErr
			continue
		}

		clusterUpdates.ToUpdate = append(clusterUpdates.ToUpdate, application)
		clusterUpdates.UpdateRef = append(clusterUpdates.UpdateRef, idx)
	}

	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterUpdates.ToUpdate))
	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionDelete, len(clusterUpdates.ToDelete))

	if len(clusterUpdates.ToUpdate) > 0 {
		_, perObjectError, innerErr := h.CreateOrUpdateObjects(clusterUpdates.ToUpdate)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
		for i, objectErr := range perObjectError {
			toIndex := clusterUpdates.UpdateRef[i]
			allObjectsErrors[toIndex] = objectErr
		}
	}

	if len(clusterUpdates.ToDelete) > 0 {
		perObjectError, innerErr := h.DeleteObjects(clusterUpdates.ToDelete)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
		for i, objectErr := range perObjectError {
			toIndex := clusterUpdates.DeleteRef[i]
			// Deleting an absent Application has already achieved the desired state.
			if !apierrors.IsNotFound(objectErr) {
				allObjectsErrors[toIndex] = objectErr
			}
		}
	}

	for idx, appSyncInfo := range response.Applications {
		h.syncReport.theReport.Updates[idx] = workloads_openapi.ClusterSyncReportUpdate{
			Id:                convutils.Ref(appSyncInfo.Application.GetId()),
			UpdateCount:       appSyncInfo.UpdateCount,
			UpdateAttemptTime: convutils.Ref(time.Now()),
		}
		if allObjectsErrors[idx] != nil {
			h.syncReport.theReport.Updates[idx].Message = *workloads_openapi.NewNullableString(convutils.Ref(allObjectsErrors[idx].Error()))
		}
	}

	return err
}

func (h *ApplicationsHandler) convertToApplication(ctx context.Context, projInfo applicationProjectInfo, item *workloads_openapi.AIApplicationSyncItem, updateCount *int64) (*runv1alpha1.Application, error) {
	app := &runv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      item.GetName(),
			Namespace: projInfo.namespace,
			Annotations: map[string]string{
				annotationApplicationUpdateCounter:                         fmt.Sprintf("%d", ptr.Deref(updateCount, 0)),
				clusterCommonConstants.ControlPlaneApplicationIdAnnotation: item.GetId(),
				annotationClusterId:                                        config.Config().ClusterUuid,
				clusterCommonConstants.AnnotationRunaiApp:                  "true",
			},
			Labels: map[string]string{
				clusterCommonConstants.DepartmentIdLabel:   projInfo.departmentId,
				clusterCommonConstants.DepartmentNameLabel: projInfo.departmentName,
				clusterCommonConstants.ProjectNameLabel:    projInfo.projectName,
				clusterCommonConstants.ProjectIdLabel:      projInfo.projectId,
			},
		},
		Spec: runv1alpha1.ApplicationSpec{
			Source: item.GetSource(),
			Type:   item.GetType(),
		},
	}

	if item.Spec != nil {
		helmSpec, credentials, err := h.convertHelmSpec(ctx, item.GetName(), projInfo.namespace, item.Spec, item.GetRequiresControlPlaneAuth())
		if err != nil {
			return nil, fmt.Errorf("failed to convert helm spec for application %s: %w", item.GetName(), err)
		}
		app.Spec.Helm = helmSpec
		if err := personalcreds.Set(app, credentials); err != nil {
			return nil, fmt.Errorf("failed to set personal credentials annotation: %w", err)
		}
	}

	return app, nil
}

func (h *ApplicationsHandler) convertHelmSpec(
	ctx context.Context,
	applicationName, namespace string,
	spec *workloads_openapi.AIApplicationSpec,
	requiresControlPlaneAuth bool,
) (*runv1alpha1.HelmSpec, []personalcreds.Cred, error) {
	helmSpec := &runv1alpha1.HelmSpec{}
	credentialSources := datastructs.NewSet[string]()

	chartRef := spec.ChartRef
	if chartRef.GetChartURL() != "" {
		helmSpec.ChartUrl = chartRef.GetChartURL()
		helmSpec.RequiresControlPlaneAuth = requiresControlPlaneAuth
	} else {
		helmSpec.ChartRef = &runv1alpha1.ChartRef{
			Repository: chartRef.GetRepo(),
			Name:       chartRef.GetChartName(),
			Version:    chartRef.GetChartVersion(),
		}
	}

	if len(spec.Values) > 0 {
		helmValues := make(map[string]apiextensionsv1.JSON, len(spec.Values))
		for key, val := range spec.Values {
			rewritten, err := rewritePersonalCredentialPlaceholders(val, applicationName, credentialSources)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to rewrite Helm values for key '%s': %w", key, err)
			}
			raw, err := json.Marshal(rewritten)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to convert Helm values for key '%s': %v", key, err)
			}
			helmValues[key] = apiextensionsv1.JSON{Raw: raw}
		}
		helmSpec.Values = helmValues
	}
	bindings, err := h.resolveCredentialBindings(ctx, applicationName, namespace, spec.CredentialBindings, credentialSources)
	if err != nil {
		return nil, nil, err
	}
	helmSpec.CredentialBindings = bindings

	auth := spec.Auth
	if auth != nil && auth.GetAssetId() != "" {
		preferredKind := preferredHelmAuthSecretKind(chartRef.GetRepo(), "")
		var secretName string
		if auth.GetUserCredential() {
			preferredKind = preferredHelmAuthSecretKind(chartRef.GetRepo(), chartRef.GetChartURL())
			sourceName, err := h.resolvePersonalSecretNameFromAssetId(ctx, auth.GetAssetId(), preferredKind)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to resolve personal chart authentication: %w", err)
			}
			credentialSources.Add(sourceName)
			secretName = personalcreds.DestName(applicationName, sourceName)
		} else {
			var err error
			secretName, err = h.resolveSecretNameFromAssetId(ctx, namespace, auth.GetAssetId(), preferredKind)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to resolve secret for asset %s in namespace %s: %w", auth.GetAssetId(), namespace, err)
			}
		}
		helmSpec.Auth = &runv1alpha1.HelmAuthSpec{
			SecretRef: &runv1alpha1.SecretReference{
				Name: secretName,
			},
		}
	}

	for sourceName := range credentialSources {
		if err := h.validatePersonalCredentialSource(ctx, sourceName); err != nil {
			return nil, nil, fmt.Errorf("invalid personal credential source: %w", err)
		}
	}

	credentials := make([]personalcreds.Cred, 0, len(credentialSources))
	sourceNames := credentialSources.Keys()
	sort.Strings(sourceNames)
	for _, sourceName := range sourceNames {
		credentials = append(credentials, personalcreds.Cred{SourceSecret: sourceName})
	}
	return helmSpec, credentials, nil
}

func rewritePersonalCredentialPlaceholders(value interface{}, applicationName string, sources datastructs.Set[string]) (interface{}, error) {
	switch typed := value.(type) {
	case string:
		if !strings.HasPrefix(typed, userCredentialPlaceholderPrefix) {
			return typed, nil
		}
		sourceName := strings.TrimPrefix(typed, userCredentialPlaceholderPrefix)
		if sourceName == "" {
			return nil, fmt.Errorf("personal credential placeholder has an empty source")
		}
		sources.Add(sourceName)
		return personalcreds.DestName(applicationName, sourceName), nil
	case map[string]interface{}:
		rewritten := make(map[string]interface{}, len(typed))
		for key, child := range typed {
			value, err := rewritePersonalCredentialPlaceholders(child, applicationName, sources)
			if err != nil {
				return nil, err
			}
			rewritten[key] = value
		}
		return rewritten, nil
	case []interface{}:
		rewritten := make([]interface{}, len(typed))
		for idx, child := range typed {
			value, err := rewritePersonalCredentialPlaceholders(child, applicationName, sources)
			if err != nil {
				return nil, err
			}
			rewritten[idx] = value
		}
		return rewritten, nil
	default:
		return value, nil
	}
}

func preferredHelmAuthSecretKind(repository, chartURL string) string {
	if strings.HasPrefix(repository, helmOCIURLPrefix) || strings.HasPrefix(chartURL, helmOCIURLPrefix) {
		return dockerRegistryKindLabel
	}
	return genericSecretKindLabel
}

func (h *ApplicationsHandler) resolveSecretNameFromAssetId(ctx context.Context, namespace string, assetId string, preferredKind ...string) (string, error) {
	secretName, err := resolveSecretName(ctx, h.cachedClient, namespace, labels.Set{
		assetIdLabel: assetId,
	}, preferredKind...)
	switch {
	case errors.Is(err, errNoMatchingSecret):
		return "", fmt.Errorf("no secret found with label %s=%s in namespace %s", assetIdLabel, assetId, namespace)
	case err != nil:
		return "", fmt.Errorf("failed to list secrets with label %s=%s in namespace %s: %w", assetIdLabel, assetId, namespace, err)
	default:
		return secretName, nil
	}
}

func (h *ApplicationsHandler) resolvePersonalSecretNameFromAssetId(ctx context.Context, assetId, preferredKind string) (string, error) {
	secretName, err := resolveSecretName(ctx, h.runaiClient, clusterCommonConstants.NamespaceRunAi, labels.Set{
		assetIdLabel:         assetId,
		userCredentialsLabel: boolTrueAsString,
	}, preferredKind)
	switch {
	case errors.Is(err, errNoMatchingSecret):
		return "", fmt.Errorf("no personal credential source found")
	case err != nil:
		return "", fmt.Errorf("failed to list personal credential sources: %w", err)
	default:
		return secretName, nil
	}
}

func resolveSecretName(
	ctx context.Context,
	reader client.Reader,
	namespace string,
	matchLabels labels.Set,
	preferredKind ...string,
) (string, error) {
	secretList := &corev1.SecretList{}
	if err := reader.List(ctx, secretList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(matchLabels),
	}); err != nil {
		return "", err
	}
	if len(secretList.Items) == 0 {
		return "", errNoMatchingSecret
	}

	if len(preferredKind) > 0 && preferredKind[0] != "" {
		for _, secret := range secretList.Items {
			if secret.Labels[clusterCommonConstants.RunaiAssetLabel] == preferredKind[0] {
				return secret.Name, nil
			}
		}
	}

	return secretList.Items[0].Name, nil
}

func (h *ApplicationsHandler) validatePersonalCredentialSource(ctx context.Context, sourceName string) error {
	source := &corev1.Secret{}
	if err := h.runaiClient.Get(ctx, types.NamespacedName{
		Namespace: clusterCommonConstants.NamespaceRunAi,
		Name:      sourceName,
	}, source); err != nil {
		return fmt.Errorf("source is unavailable: %w", err)
	}
	if source.Labels[userCredentialsLabel] != boolTrueAsString {
		return fmt.Errorf("source is not labelled as a personal credential")
	}
	return nil
}

func (h *ApplicationsHandler) getBackendApplications(updatesAfter int64) (*workloads_openapi.AIApplicationSyncResponse, error) {
	request := h.backendClient.WorkloadsClient.AIApplicationsAPI.RequestApplicationsUpdates(h.backendClient.Context())
	request = request.ClusterId(config.Config().ClusterUuid)
	request = request.UpdatesAfter(updatesAfter)
	if h.syncReport != nil && h.syncReport.numApplications > 0 {
		request = request.ClusterSyncReport(h.syncReport.theReport)
	}

	syncResponse, res, err := request.Execute()
	if res != nil && res.Body != nil {
		defer func() { _ = res.Body.Close() }()
	}
	if err != nil {
		log.Error().Msgf("Had an error while preparing applications get request: %s", err.Error())
		return nil, err
	}

	err = h.backendClient.HandleHttpResponse(res, err, "get applications updates")
	if err != nil {
		log.Error().Msgf("Had an error while getting backend applications: %s", err.Error())
		return nil, err
	}

	return syncResponse, nil
}

func (h *ApplicationsHandler) setSyncCounter(highestUpdate int64) error {
	err := configmaps.UpdateApplicationsSyncCounter(h.runaiClient, h.cachedClient, highestUpdate)
	if err != nil {
		return err
	}
	h.syncCounter = highestUpdate
	log.Info().Msgf("Updated applications sync counter to: %d", highestUpdate)
	return nil
}

type ApplicationSpecCopier struct{}

func (c *ApplicationSpecCopier) DeepCopySpec(from, to client.Object) error {
	fromApp, ok := from.(*runv1alpha1.Application)
	if !ok {
		return fmt.Errorf("from object is not an Application")
	}

	toApp, ok := to.(*runv1alpha1.Application)
	if !ok {
		return fmt.Errorf("to object is not an Application")
	}

	toApp.Spec = *fromApp.Spec.DeepCopy()
	return nil
}

type ApplicationObjectInspector struct {
	specCopier *ApplicationSpecCopier
}

func (a ApplicationObjectInspector) DeepCopySpec(from, to client.Object) {
	err := a.specCopier.DeepCopySpec(from, to)
	if err != nil {
		log.Error().Msgf("Had an error while copying application spec: %s", err.Error())
		return
	}

	if to.GetAnnotations() == nil {
		to.SetAnnotations(map[string]string{})
	}
	to.SetAnnotations(from.GetAnnotations())
}

func (a ApplicationObjectInspector) IsUpToDate(left, right client.Object) bool {
	if left == nil || right == nil {
		return false
	}

	rightAnnotations := right.GetAnnotations()
	counterStr, ok := rightAnnotations[annotationApplicationUpdateCounter]
	if !ok || counterStr == "" {
		return false
	}
	rightCounter, err := strconv.Atoi(counterStr)
	if err != nil {
		log.Warn().Msgf("Had an error while converting application update counter annotation: %s, err: %s", counterStr, err.Error())
		return false
	}

	leftAnnotations := left.GetAnnotations()
	counterStr, ok = leftAnnotations[annotationApplicationUpdateCounter]
	if !ok || counterStr == "" {
		return false
	}
	leftCounter, err := strconv.Atoi(counterStr)
	if err != nil {
		log.Warn().Msgf("Had an error while converting application update counter annotation: %s, err: %s", counterStr, err.Error())
		return false
	}

	return leftCounter <= rightCounter
}

func NewApplicationsHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client, cachedClient client.Reader) (ResourceHandler, error) {
	object := &runv1alpha1.Application{}
	objectInspector := ApplicationObjectInspector{
		specCopier: &ApplicationSpecCopier{},
	}

	handlingOptions := HandlingOptions{deletePropagationOrphan: false}
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient, object, objectInspector, handlingOptions)
	if err != nil {
		return nil, err
	}

	agentConfigData, _ := configmaps.GetAgentConfig(cachedClient)

	return &ApplicationsHandler{
		CommonHandler: commonHandler,
		syncCounter:   agentConfigData.ApplicationsSyncCounter,
	}, nil
}
