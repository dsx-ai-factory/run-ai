package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	workloads_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	runaibackendclient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	workloadtypeconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/workload-type-config"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	WorkloadsTypesConfigMetricLabel = "workloads_config"
)

// WorkloadsConfigHandler uses two handlers since the backend returns two data sets:
// 1) Categories and workload types stored in a ConfigMap (cmCommonHandler)
// 2) Karta CRDs parsed from types with Create/Delete (kartaCommonHandler)
type WorkloadsConfigHandler struct {
	cmCommonHandler    *CommonHandler
	kartaCommonHandler *WorkloadTypeKartasHandler
}

func NewWorkloadsConfigHandler(runaiBackendClient *runaibackendclient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	cmCommonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient, GetObjectType(), WtcConfigMapObjectInspector{}, HandlingOptions{})
	if err != nil {
		return nil, err
	}
	kartaCommonHandler, err := NewWorkloadTypeKartasHandler(runaiBackendClient, runaiClient, cachedClient)
	if err != nil {
		return nil, err
	}
	return &WorkloadsConfigHandler{
		cmCommonHandler:    cmCommonHandler,
		kartaCommonHandler: kartaCommonHandler,
	}, nil
}

func (w *WorkloadsConfigHandler) Handle() error {
	categoriesResponse, err := w.getWorkloadsCategories()
	if err != nil {
		return err
	}
	configsResponse, err := w.getWorkloadsConfigs()
	if err != nil {
		return err
	}

	errInner := w.handleCm(categoriesResponse, configsResponse)
	if errInner != nil {
		err = multierror.Append(err, errInner)
	}

	errInner = w.handleKartas(configsResponse)
	if errInner != nil {
		err = multierror.Append(err, errInner)
	}

	return err
}

type WtcConfigMapObjectInspector struct{}

type WtcKartaObjectInspector struct{}

func (w *WorkloadsConfigHandler) MetricLabel() string {
	return WorkloadsTypesConfigMetricLabel
}

func getEmptyObject() *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	cm.Name = constants.WorkloadsConfigMapName
	cm.Namespace = constants.WorkloadsConfigMapNamespace
	return cm
}

func convertToClusterWorkloadTypeConfig(configs []workloads_openapi.WorkloadTypeConfig) []workloadtypeconfig.WorkloadTypeConfig {
	var wc []workloadtypeconfig.WorkloadTypeConfig
	for _, config := range configs {
		wc = append(wc, workloadtypeconfig.WorkloadTypeConfig{
			TypeName:       config.Name,
			Group:          config.Group,
			CategoryName:   ptr.Deref(config.CategoryName, ""),
			PriorityName:   ptr.Deref(config.PriorityName, ""),
			Preemptibility: string(config.Preemptibility),
		})
	}
	return wc
}

func convertCategoryToString(configs []workloads_openapi.WorkloadCategory) []string {
	categoryNames := []string{}
	for _, workloadCategory := range configs {
		categoryNames = append(categoryNames, workloadCategory.Name)
	}
	return categoryNames
}

func setWorkloadsTypeConfig(cm *corev1.ConfigMap, configs []workloads_openapi.WorkloadTypeConfig) error {
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	wc := convertToClusterWorkloadTypeConfig(configs)
	jsonData, err := json.Marshal(wc)
	if err != nil {
		return fmt.Errorf("failed to marshal workload type configs: %w", err)
	}

	cm.Data[constants.WorkloadsConfigMapTypesKey] = string(jsonData)
	return nil
}

func setWorkloadsCategories(cm *corev1.ConfigMap, categories []workloads_openapi.WorkloadCategory) error {
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	names := convertCategoryToString(categories)
	jsonData, err := json.Marshal(names)
	if err != nil {
		return fmt.Errorf("failed to marshal workload categories: %w", err)
	}

	cm.Data[constants.WorkloadsConfigMapCategoriesKey] = string(jsonData)
	return nil
}

func (w WtcConfigMapObjectInspector) IsUpToDate(left, right client.Object) bool {
	if left == nil || right == nil {
		return false
	}

	leftCM, ok := left.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	rightCM, ok := right.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	leftTypesData, leftTypesExists := leftCM.Data[constants.WorkloadsConfigMapTypesKey]
	rightTypesData, rightTypesExists := rightCM.Data[constants.WorkloadsConfigMapTypesKey]

	leftCategoriesData, leftCategoriesExists := leftCM.Data[constants.WorkloadsConfigMapCategoriesKey]
	rightCategoriesData, rightCategoriesExists := rightCM.Data[constants.WorkloadsConfigMapCategoriesKey]

	if !leftTypesExists || !rightTypesExists || !leftCategoriesExists || !rightCategoriesExists {
		return false
	}

	return leftTypesData == rightTypesData && leftCategoriesData == rightCategoriesData
}

func (w WtcConfigMapObjectInspector) DeepCopySpec(from, to client.Object) {
	fromCM, ok := from.(*corev1.ConfigMap)
	if !ok {
		return
	}

	toCM, ok := to.(*corev1.ConfigMap)
	if !ok {
		return
	}

	if toCM.Data == nil {
		toCM.Data = make(map[string]string)
	}

	if data, exists := fromCM.Data[constants.WorkloadsConfigMapTypesKey]; exists {
		toCM.Data[constants.WorkloadsConfigMapTypesKey] = data
	}

	if data, exists := fromCM.Data[constants.WorkloadsConfigMapCategoriesKey]; exists {
		toCM.Data[constants.WorkloadsConfigMapCategoriesKey] = data
	}
}

func (w WtcKartaObjectInspector) IsUpToDate(expected, existing client.Object) bool {
	// Kartas have no update method in the control-plane, but could be updated by us as a db migration.
	// Therefore, we only need to validate labels -
	// one of the labels should be "updatedAt" which changes when the backend Karta changes.
	expectedLabels := expected.GetLabels()
	existingLabels := existing.GetLabels()

	for expectedKey, expectedValue := range expectedLabels {
		actualValue, exists := existingLabels[expectedKey]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	return true
}

func (w WtcKartaObjectInspector) DeepCopySpec(from, to client.Object) {
	fromKarta, ok := from.(*kartav1alpha1.Karta)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toKarta, ok := to.(*kartav1alpha1.Karta)
	if !ok {
		return
	}

	fromKarta.Spec.DeepCopyInto(&toKarta.Spec)

	if toKarta.Labels == nil {
		toKarta.Labels = make(map[string]string)
	}
	for expectedKey, expectedValue := range fromKarta.Labels {
		toKarta.Labels[expectedKey] = expectedValue
	}
}

func GetObjectType() client.Object {
	return &corev1.ConfigMap{}
}

func (w *WorkloadsConfigHandler) handleCm(categoriesResponse *workloads_openapi.ListCategories200Response, configsResponse *workloads_openapi.ListWorkloadTypes200Response) error {
	cm := getEmptyObject()
	err := setWorkloadsTypeConfig(cm, configsResponse.Types)
	if err != nil {
		return fmt.Errorf("failed to set workloads type config map: %w", err)
	}

	// Set workload categories
	err = setWorkloadsCategories(cm, categoriesResponse.Categories)
	if err != nil {
		return fmt.Errorf("failed to set workloads type config map: %w", err)
	}

	// Update the config map
	err = w.cmCommonHandler.CreateOrUpdateObject(cm, "")
	if err != nil {
		return fmt.Errorf("failed to update workloads config map: %w", err)
	}

	return nil
}

func (w *WorkloadsConfigHandler) handleKartas(configsResponse *workloads_openapi.ListWorkloadTypes200Response) error {
	log.Debug().Msgf("Starting to sync Kartas")

	clusterKartas, extractErr := w.kartaCommonHandler.extractKartas(configsResponse)
	if extractErr != nil {
		log.Error().Msgf("Failed to extract some Kartas, partial=%d, error: %s", len(clusterKartas), extractErr.Error())
	}

	metrics.GetMetricsExporter().AddResourceAction(w.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterKartas))

	reconcileErr := w.kartaCommonHandler.ReconcileObjects(clusterKartas, extractErr != nil)
	if extractErr != nil {
		return multierror.Append(reconcileErr, extractErr)
	}
	return reconcileErr
}

func (w *WorkloadsConfigHandler) getWorkloadsConfigs() (*workloads_openapi.ListWorkloadTypes200Response, error) {
	var configsResponse *workloads_openapi.ListWorkloadTypes200Response
	var res *http.Response
	var err error

	request := w.cmCommonHandler.backendClient.WorkloadsClient.WorkloadPropertiesAPI.ListWorkloadTypes(w.cmCommonHandler.backendClient.Context())
	request = request.NoStatus(true)
	configsResponse, _, err = request.Execute()
	err = w.cmCommonHandler.backendClient.HandleHttpResponse(res, err, "get workloads types config")

	if err != nil {
		log.Error().Msgf("Had an error while getting backend workload: %s", err.Error())
		return nil, err
	}

	return configsResponse, nil
}

func (w *WorkloadsConfigHandler) getWorkloadsCategories() (*workloads_openapi.ListCategories200Response, error) {
	var categoriesResponse *workloads_openapi.ListCategories200Response
	var res *http.Response
	var err error

	request := w.cmCommonHandler.backendClient.WorkloadsClient.WorkloadPropertiesAPI.ListCategories(w.cmCommonHandler.backendClient.Context())
	categoriesResponse, _, err = request.Execute()
	err = w.cmCommonHandler.backendClient.HandleHttpResponse(res, err, "get workloads categories")

	if err != nil {
		log.Error().Msgf("Had an error while getting backend workload: %s", err.Error())
		return nil, err
	}

	return categoriesResponse, nil
}

func extractKarta(
	apiKarta map[string]interface{},
	updatedAt *time.Time,
	version,
	workloadTypeName,
	workloadTypeGroup,
	workloadTypeId string,
	onlyMetadata bool) (client.Object, error) {
	kartaObj := &kartav1alpha1.Karta{}

	if !onlyMetadata {
		jsonBytes, err := json.Marshal(apiKarta)
		if err != nil {
			return nil, fmt.Errorf("failed to json marshal Karta of workloadTypeConfig <%s/%s>, error: %s",
				workloadTypeName, workloadTypeGroup, err.Error())
		}

		err = json.Unmarshal(jsonBytes, &kartaObj)
		if err != nil {
			return nil, fmt.Errorf("failed to json un-marshal Karta of workloadTypeConfig <%s/%s>, error: %s",
				workloadTypeName, workloadTypeGroup, err.Error())
		}
	}

	updatedAtStr := ""
	if updatedAt != nil {
		updatedAtStr = strings.ReplaceAll(updatedAt.Format(time.RFC3339), ":", "-")
	}

	name, err := v1alpha1.GenerateKartaNameByGvk(workloadTypeGroup, version, workloadTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate name for Karta of workloadTypeConfig <%s/%s>, error: %s",
			workloadTypeName, workloadTypeGroup, err.Error())
	}
	kartaObj.SetName(name)

	kartaObj.SetLabels(v1alpha1.GetKartaLabels(
		workloadTypeId,
		workloadTypeGroup,
		version,
		workloadTypeName,
		updatedAtStr,
	))

	return kartaObj, nil
}

// generateKartaName extracts the GVK from a Karta and delegates to the
// shared SDK helper so agent and karta-migrate produce identical names.
func generateKartaName(k *kartav1alpha1.Karta) (string, error) {
	if k.Spec.StructureDefinition.RootComponent.Kind == nil {
		return "", fmt.Errorf("karta kind is nil")
	}
	kind := k.Spec.StructureDefinition.RootComponent.Kind.Kind
	group := k.Spec.StructureDefinition.RootComponent.Kind.Group
	version := k.Spec.StructureDefinition.RootComponent.Kind.Version
	return v1alpha1.GenerateKartaNameByGvk(group, version, kind)
}

type kartaForAgent struct {
	apiKarta          map[string]interface{}
	updatedAt         *time.Time
	version           string
	workloadTypeName  string
	workloadTypeGroup string
	workloadTypeId    string
}

// WorkloadTypeKartasHandler handles the kartas of a workload type;
// to override some funcs.
type WorkloadTypeKartasHandler struct {
	*CommonHandler
	backendWorkloadTypes map[string]kartaForAgent
}

func NewWorkloadTypeKartasHandler(runaiBackendClient *runaibackendclient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (*WorkloadTypeKartasHandler, error) {
	kartaCommonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient, &kartav1alpha1.Karta{}, WtcKartaObjectInspector{}, HandlingOptions{})
	if err != nil {
		return nil, err
	}

	workloadTypeKartasHandler := &WorkloadTypeKartasHandler{
		CommonHandler:        kartaCommonHandler,
		backendWorkloadTypes: map[string]kartaForAgent{},
	}
	workloadTypeKartasHandler.SetExpectedObjectGetter(workloadTypeKartasHandler)

	return workloadTypeKartasHandler, nil
}

func (h *WorkloadTypeKartasHandler) GetExpectedObject(name, _ string) (client.Object, error) {
	// only json-parse the object once, here, when needed to create it
	backendKarta, ok := h.backendWorkloadTypes[name]
	if !ok {
		return nil, fmt.Errorf("failed to find backend karta for <%s>", name)
	}

	kartaObj, err := extractKarta(
		backendKarta.apiKarta,
		backendKarta.updatedAt,
		backendKarta.version,
		backendKarta.workloadTypeName,
		backendKarta.workloadTypeGroup,
		backendKarta.workloadTypeId,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract karta for <%s>, error: %s", name, err.Error())
	}
	return kartaObj, nil
}

// ShouldReCreate checks if a workload type karta should be recreated.
// In case there was a change in the labels (see 'IsUpToDate'),
// this func should be called with the full objects to check if the change requires re-creation of the object.
func (h *WorkloadTypeKartasHandler) ShouldReCreate(expected, existing client.Object) bool {
	expectedKarta, ok := expected.(*kartav1alpha1.Karta)
	if !ok {
		return false
	}
	existingKarta, ok := existing.(*kartav1alpha1.Karta)
	if !ok {
		return false
	}

	setKartaDefaultsToAvoidRecreation(expectedKarta, existingKarta)

	// if the specs are not equal - means the Karta changed in the backend - need to re-create it.
	// else - only the labels changed - no need to re-create.
	return !reflect.DeepEqual(expectedKarta.Spec, existingKarta.Spec)
}

func setKartaDefaultsToAvoidRecreation(expectedKarta, existingKarta *kartav1alpha1.Karta) {
	if expectedKarta.Spec.StructureDefinition.RootComponent.StatusDefinition != nil &&
		expectedKarta.Spec.StructureDefinition.RootComponent.StatusDefinition.ConditionsDefinition != nil &&
		expectedKarta.Spec.StructureDefinition.RootComponent.StatusDefinition.ConditionsDefinition.MessageFieldName == nil {
		// to avoid unnecessary re-creation due to defaulting of the MessageFieldName to "message"
		if existingKarta.Spec.StructureDefinition.RootComponent.StatusDefinition != nil &&
			existingKarta.Spec.StructureDefinition.RootComponent.StatusDefinition.ConditionsDefinition != nil {
			existingKarta.Spec.StructureDefinition.RootComponent.StatusDefinition.ConditionsDefinition.MessageFieldName = nil
		}
	}
}

func (h *WorkloadTypeKartasHandler) extractKartas(configsResponse *workloads_openapi.ListWorkloadTypes200Response) ([]client.Object, error) {
	h.backendWorkloadTypes = map[string]kartaForAgent{}

	clusterKartas := []client.Object{}
	var err error

	for _, wtc := range configsResponse.Types {
		for _, kartaInner := range wtc.Kartas {
			if kartaInner.Karta == nil {
				continue
			}

			group := wtc.GetGroup()
			version := kartaInner.GetVersion()
			kind := wtc.GetName()
			workloadTypeId := wtc.GetId()

			// create an object with only name, gvk and labels, to be used in the reconciliation process
			kartaObj, innerErr := extractKarta(
				kartaInner.Karta,
				kartaInner.UpdatedAt,
				version,
				kind,
				group,
				workloadTypeId,
				true,
			)
			if innerErr != nil {
				err = multierror.Append(err, innerErr)
				continue
			}

			h.backendWorkloadTypes[kartaObj.GetName()] = kartaForAgent{
				apiKarta:          kartaInner.Karta,
				updatedAt:         kartaInner.UpdatedAt,
				version:           version,
				workloadTypeName:  kind,
				workloadTypeGroup: group,
				workloadTypeId:    workloadTypeId,
			}

			clusterKartas = append(clusterKartas, kartaObj)
		}
	}

	return clusterKartas, err
}
