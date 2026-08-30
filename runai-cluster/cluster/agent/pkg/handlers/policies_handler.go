package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/common/utils"

	"github.com/run-ai/runai/runai-application-packages/concurrency"
	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	policy_openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	configmaps "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config-maps"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	infpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxNumWorkers = 4

	_policySuffix = "Policy" //nolint:unused
)

// PoliciesSyncReportInfo is the report to be sent to the agent on  each sync request
// reporting success/failure of the previous sync cycle. it is not a must-sent
// kind of information, but it helps the control plane in displaying failure/success
type PoliciesSyncReportInfo struct {
	numPolicies int
	theReport   policy_openapi.ClusterSyncReport
}

var syncReportPerPolicyType = concurrency.NewConcurrentMap[*PoliciesSyncReportInfo](nil)

type PoliciesHandler struct {
	*CommonHandler
	syncCounter int64
	policyType  string
	metricLabel string
}

func (p *PoliciesHandler) MetricLabel() string {
	return p.metricLabel
}

type PoliciesObjectInspector bool

func (p PoliciesObjectInspector) DeepCopySpec(from, to client.Object) {
	// spec & status
	switch to.GetObjectKind().GroupVersionKind().Kind {
	case common.PolicyTypeInteractive:
		fromPolicy, ok := from.(*ipv1.InteractivePolicy)
		if !ok {
			return
		}
		toPolicy, ok := to.(*ipv1.InteractivePolicy)
		if !ok {
			return
		}
		fromPolicy.Spec.DeepCopyInto(&toPolicy.Spec)
		fromPolicy.Status.DeepCopyInto(&toPolicy.Status)
	case common.PolicyTypeTraining:
		fromPolicy, ok := from.(*tpv1.TrainingPolicy)
		if !ok {
			return
		}
		toPolicy, ok := to.(*tpv1.TrainingPolicy)
		if !ok {
			return
		}
		fromPolicy.Spec.DeepCopyInto(&toPolicy.Spec)
		fromPolicy.Status.DeepCopyInto(&toPolicy.Status)
	case common.PolicyTypeDistributed:
		fromPolicy, ok := from.(*dpv1.DistributedPolicy)
		if !ok {
			return
		}
		toPolicy, ok := to.(*dpv1.DistributedPolicy)
		if !ok {
			return
		}
		fromPolicy.Spec.DeepCopyInto(&toPolicy.Spec)
		if fromPolicy.MasterSpec != nil {
			if toPolicy.MasterSpec == nil {
				toPolicy.MasterSpec = &dpv1.DistributedMasterSpec{}
			}
			fromPolicy.MasterSpec.DeepCopyInto(toPolicy.MasterSpec)
		}
		fromPolicy.Status.DeepCopyInto(&toPolicy.Status)
	case common.PolicyTypeInference:
		fromPolicy, ok := from.(*infpv1.InferencePolicy)
		if !ok {
			return
		}
		toPolicy, ok := to.(*infpv1.InferencePolicy)
		if !ok {
			return
		}
		fromPolicy.Spec.DeepCopyInto(&toPolicy.Spec)
		fromPolicy.Status.DeepCopyInto(&toPolicy.Status)
	}

	// annotations
	if to.GetAnnotations() == nil {
		to.SetAnnotations(map[string]string{})
	}
	to.SetAnnotations(from.GetAnnotations())
}

func (p PoliciesObjectInspector) IsUpToDate(left, right client.Object) bool {
	if left == nil || right == nil {
		return false
	}
	rightAnnotations := right.GetAnnotations()
	rightAnnotationsCounter, err := strconv.Atoi(rightAnnotations[constants.AnnotationsBackendPolicyCounter])
	if err != nil {
		log.Warn().Msgf("Had an error while converting backend policy counter to int object annotation: %s , err: %s", rightAnnotations[constants.AnnotationsBackendPolicyCounter], err.Error())
		return false
	}

	leftAnnotations := left.GetAnnotations()
	leftAnnotationsCounter, err := strconv.Atoi(leftAnnotations[constants.AnnotationsBackendPolicyCounter])
	if err != nil {
		log.Warn().Msgf("Had an error while converting backend policy counter to int object annotation: %s , err: %s", leftAnnotations[constants.AnnotationsBackendPolicyCounter], err.Error())
		return false
	}
	if leftAnnotationsCounter <= rightAnnotationsCounter {
		// if the backend policy counter is the same or higher than the cluster policy counter,
		// it means that the backend policy is newer than the cluster policy
		return true
	}
	return false
}

func (p *PoliciesHandler) Handle() error {
	var err error
	var prevHighestUpdate int64
	numWorkers := maxNumWorkers

	//
	//  make sure each policy type has an entry in the sync report
	//
	syncReport, found := syncReportPerPolicyType.Load(p.policyType)
	if !found {
		syncReport = &PoliciesSyncReportInfo{}
		syncReportPerPolicyType.Store(p.policyType, syncReport)
	}

	for {
		backendPolicies, innerErr := p.getBackendPolicies(p.syncCounter)
		if innerErr != nil {
			log.Error().Msgf("Had an error while getting backend policies, sync counter: %d, error: %s", p.syncCounter, innerErr.Error())
			return innerErr
		}

		var highestUpdate, highestSysUpdate int64
		if backendPolicies != nil && backendPolicies.ResponseInfo != nil && backendPolicies.ResponseInfo.HighestUpdate != nil {
			highestUpdate = *backendPolicies.ResponseInfo.HighestUpdate
		}

		if backendPolicies != nil && backendPolicies.SystemInfo != nil && backendPolicies.SystemInfo.HighestUpdate != nil {
			highestSysUpdate = *backendPolicies.SystemInfo.HighestUpdate
		}

		numPoliciesToSync := len(backendPolicies.Policies)
		syncReport.numPolicies = numPoliciesToSync

		if numPoliciesToSync == 0 {
			log.Debug().Msgf("Backend %s policies response - prev highest update: %d, highest update: %d, highest system update: %d, size: %d", p.policyType, prevHighestUpdate, highestUpdate, highestSysUpdate, numPoliciesToSync)
			log.Debug().Msgf("No %s policies to handle while counter is %d", p.policyType, p.syncCounter)
			break
		}
		log.Info().Msgf("Backend %s policies response - prev highest update: %d, highest update: %d, highest system update: %d, size: %d", p.policyType, prevHighestUpdate, highestUpdate, highestSysUpdate, numPoliciesToSync)

		syncReport.theReport.Updates = make([]policy_openapi.ClusterSyncReportUpdate, numPoliciesToSync)

		innerErr = p.ProcessPoliciesConcurrently(numPoliciesToSync, numWorkers, backendPolicies)
		if innerErr != nil {
			log.Error().Msgf("Had an error while processing %s policies, error: %s", p.policyType, innerErr.Error())
			err = multierror.Append(err, innerErr)
		}

		// update the highest update counter
		innerErr = p.setSyncCounter(highestUpdate)
		if innerErr != nil {
			log.Error().Msgf("Had an error while updating %s policies sync counter, error: %s", p.policyType, innerErr.Error())
			err = multierror.Append(err, innerErr)
		}

		// if the highest update counter is the same as the highest system update counter, it means that there are no more cluster policies to handle
		if highestUpdate == highestSysUpdate {
			log.Debug().Msgf("No more %s policies to handle", p.policyType)
			break
		}
	}

	if err != nil {
		log.Error().Msgf("Failed to handle some of sync policies (create, update, or delete), error: %s", err.Error())
		return err
	}

	log.Debug().Msgf("Finished handled cluster Policies")
	return err
}

func (p *PoliciesHandler) ProcessPoliciesConcurrently(numPoliciesToSync int, numWorkers int, backendPolicies *policy_openapi.PolicySyncResponse) error {
	var wg sync.WaitGroup
	var err error
	var errMu sync.Mutex

	// calc and update metric
	deletedCount := 0
	for _, policy := range backendPolicies.Policies {
		if policy.IsDeleted.IsSet() {
			deletedCount++
		}
	}
	metrics.GetMetricsExporter().AddResourceAction(p.MetricLabel(), metrics.ActionCreateOrUpdate, len(backendPolicies.Policies)-deletedCount)
	metrics.GetMetricsExporter().AddResourceAction(p.MetricLabel(), metrics.ActionDelete, deletedCount)

	syncReport, _ := syncReportPerPolicyType.Load(p.policyType)

	chunkSize := numPoliciesToSync / numWorkers

	if chunkSize == 0 {
		// if there are less policies than workers, we will have only one worker who will handle all policies
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		// Calculate the start and end indices for each worker's chunk
		start := i * chunkSize
		end := (i + 1) * chunkSize

		// The last worker may get extra items if clusterPoliciesSize is not divisible by numWorkers
		if i == numWorkers-1 {
			end = numPoliciesToSync
		}

		log.Debug().Msgf("Worker %d: start processing, start: %v, end: %v", i, start, end)
		go func() {
			// wg.Done is intentionally deferred here (and not inside ProcessPolicies)
			// so that wg.Wait only returns after the report-writing loop below finishes.
			// Otherwise the caller could reallocate syncReport.theReport.Updates (which
			// is stored in a package-level global keyed by policy type) before this
			// goroutine finishes writing to it, causing an index-out-of-range panic.
			defer wg.Done()

			perWorkloadError, innerErr := p.ProcessPolicies(backendPolicies.Policies[start:end])
			if innerErr != nil {
				errMu.Lock()
				err = multierror.Append(err, innerErr)
				errMu.Unlock()
			}

			for idx := start; idx < end; idx++ {
				syncReport.theReport.Updates[idx] = policy_openapi.ClusterSyncReportUpdate{
					Id:                convutils.Ref(backendPolicies.Policies[idx].Meta.Name),
					UpdateCount:       backendPolicies.Policies[idx].UpdateCounter,
					UpdateAttemptTime: convutils.Ref(time.Now()),
				}
				if perWorkloadError[idx-start] != nil {
					syncReport.theReport.Updates[idx].Message = *policy_openapi.NewNullableString(utils.Ref(perWorkloadError[idx-start].Error()))
				}
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	return err
}

func (p *PoliciesHandler) ProcessPolicies(policiesChunk []policy_openapi.PolicySyncResponsePoliciesInner) ([]error, error) {
	var err error

	allObjectsErrors := make([]error, len(policiesChunk))

	policiesToHandle, innerErr := p.backendToClusterWorkspacePolicies(policiesChunk, allObjectsErrors)
	if innerErr != nil {
		log.Error().Msgf("Had an error while getting backend policies: %s", innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	if len(policiesToHandle.ToUpdate) > 0 {
		_, perObjectErr, innerErr := p.CreateOrUpdateObjects(policiesToHandle.ToUpdate)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		for idx, objectErr := range perObjectErr {
			toIndex := policiesToHandle.UpdateRef[idx]
			allObjectsErrors[toIndex] = objectErr
		}
	}

	if len(policiesToHandle.ToDelete) > 0 {
		perObjectErr, innerErr := p.DeleteObjects(policiesToHandle.ToDelete)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}

		for idx, objectErr := range perObjectErr {
			toIndex := policiesToHandle.DeleteRef[idx]
			allObjectsErrors[toIndex] = objectErr
		}
	}

	return allObjectsErrors, err
}

func (p *PoliciesHandler) setSyncCounter(highestUpdate int64) error {
	p.syncCounter = highestUpdate
	switch p.policyType {
	case common.PolicyTypeInteractive:
		err := configmaps.UpdateInteractivePoliciesSyncCounters(p.runaiClient, p.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated interactive policies sync counter to: %d", highestUpdate)
	case common.PolicyTypeTraining:
		err := configmaps.UpdateTrainingPoliciesSyncCounters(p.runaiClient, p.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated training policies sync counter to: %d", highestUpdate)
	case common.PolicyTypeDistributed:
		err := configmaps.UpdateDistributedPoliciesSyncCounters(p.runaiClient, p.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated distributed policies sync counter to: %d", highestUpdate)
	case common.PolicyTypeInference:
		err := configmaps.UpdateInferencePoliciesSyncCounters(p.runaiClient, p.cachedClient, highestUpdate)
		if err != nil {
			return err
		}
		log.Info().Msgf("Updated inference policies sync counter to: %d", highestUpdate)
	}
	return nil
}

func (p *PoliciesHandler) backendToClusterWorkspacePolicies(policies []policy_openapi.PolicySyncResponsePoliciesInner,
	perPolicyErr []error) (clusterWorkloadsAndPoliciesUpdates, error) {

	projectToNamespaceList, err := p.ProjectToNamespaceList()
	if err != nil {
		return clusterWorkloadsAndPoliciesUpdates{}, err
	}

	clusterPoliciesExtraction := clusterWorkloadsAndPoliciesUpdates{}

	for idx, backendPolicy := range policies {
		backendPolicyCounter := *backendPolicy.UpdateCounter
		backendPolicyNamespace := constants.NamespaceRunAi // default namespace for tenant/department scope

		if backendPolicy.Meta.Scope == policy_openapi.SCOPETYPE_PROJECT {
			projectIdPtr := backendPolicy.Meta.ProjectId.Get()
			if projectIdPtr != nil {
				var found bool
				projectId := strconv.Itoa(int(*projectIdPtr))
				backendPolicyNamespace, found = projectToNamespaceList[projectId]
				if !found || backendPolicyNamespace == "" {
					log.Warn().Msgf("Had an error while getting backend policy namespace, project id: %s", projectId)
					continue
				}
			} else {
				log.Warn().Msgf("Had an error while getting backend policy project id, name: %s, project id: %d", backendPolicy.Meta.Name, *projectIdPtr)
				continue
			}
		}

		// filter the policy if there is a conflict with a "first generation"-"cluster" policy
		if p.overloadingWithAppliedClusterPolicy(backendPolicy.Meta.Name, backendPolicyNamespace) {
			log.Error().Msgf("Policy %s:%s has a conflict with already existing policy is a cluster, "+
				"agent doesn't create/update/delete cluster policy. it should first be deleted manually",
				backendPolicy.Meta.Name, backendPolicyNamespace)
			continue
		}

		if backendPolicy.IsDeleted.IsSet() && *backendPolicy.IsDeleted.Get() {
			clusterPoliciesExtraction.ToDelete = append(clusterPoliciesExtraction.ToDelete, types.NamespacedName{Name: backendPolicy.Meta.Name, Namespace: backendPolicyNamespace})
			clusterPoliciesExtraction.DeleteRef = append(clusterPoliciesExtraction.DeleteRef, idx)
			continue
		}

		// the context is used for loading cluster information in the control plane, in agent we do not need it
		clusterPolicy, err := p.composePolicy(context.Background(), &backendPolicy, backendPolicyNamespace, backendPolicyCounter)
		if err != nil {
			formattedError := err.Error()
			log.Error().Msgf("Had an error while unmarshalling backend policy: %s", formattedError)
			perPolicyErr[idx] = err
			continue
		}

		clusterPoliciesExtraction.ToUpdate = append(clusterPoliciesExtraction.ToUpdate, clusterPolicy)
		clusterPoliciesExtraction.UpdateRef = append(clusterPoliciesExtraction.UpdateRef, idx)
	}

	return clusterPoliciesExtraction, nil
}

// overloadingWithAppliedClusterPolicy check if there is already a "first generation" policy with the same name and namespace
func (p *PoliciesHandler) overloadingWithAppliedClusterPolicy(name string, namespace string) bool {
	// Check if there is an existing policy with the same name and namespace
	existingObject, err := p.getObject(name, namespace)
	if err != nil {
		// This function mitigate risks, therefore preferred action is to continue the flow in case of an error
		return false
	}

	// "first generation" policy are detected by NOT having the backend policy counter annotation
	annotations := existingObject.GetAnnotations()
	_, ok := annotations[constants.AnnotationsBackendPolicyCounter]
	return !ok
}

func (p *PoliciesHandler) composePolicy(_ context.Context, backendUpdate *policy_openapi.PolicySyncResponsePoliciesInner,
	backendPolicyNamespace string, backendUpdateCount int64) (client.Object, error) {

	log.Error().Msgf("PolicySpec is no longer generated by the agent, handle object meta and status only")

	switch p.policyType {
	case common.PolicyTypeInteractive:
		policy := ipv1.InteractivePolicy{
			ObjectMeta: p.initObjectMeta(&backendUpdate.Meta, backendPolicyNamespace, backendUpdateCount),
			Status:     p.initStatus(backendUpdateCount),
		}
		return &policy, nil

	case common.PolicyTypeTraining:
		policy := tpv1.TrainingPolicy{
			ObjectMeta: p.initObjectMeta(&backendUpdate.Meta, backendPolicyNamespace, backendUpdateCount),
			Status:     p.initStatus(backendUpdateCount),
		}
		return &policy, nil

	case common.PolicyTypeDistributed:
		policy := dpv1.DistributedPolicy{
			ObjectMeta: p.initObjectMeta(&backendUpdate.Meta, backendPolicyNamespace, backendUpdateCount),
			Status:     p.initStatus(backendUpdateCount),
		}
		return &policy, nil

	case common.PolicyTypeInference:
		policy := infpv1.InferencePolicy{
			ObjectMeta: p.initObjectMeta(&backendUpdate.Meta, backendPolicyNamespace, backendUpdateCount),
			Status:     p.initStatus(backendUpdateCount),
		}
		return &policy, nil

	default:
		return nil, fmt.Errorf("unsupported policy type: %s", p.policyType)
	}
}

func (p *PoliciesHandler) _loadDefaultsAndRules(effective policy_openapi.NullableSupersetPolicy) (*policy_openapi.SupersetDefaults, error) { //nolint:unused

	if !effective.IsSet() {
		return nil, fmt.Errorf("failed to get effective policy for sync")
	}

	var defaults policy_openapi.SupersetDefaults

	errCopy := _copyDefaultsOrRules(&defaults, effective.Get().GetDefaults())
	if errCopy != nil {
		return nil, fmt.Errorf("failed to load defaults of policy: %v", errCopy)
	}

	return &defaults, nil
}

func (p *PoliciesHandler) _loadDistributedDefaultsAndRules(distributed policy_openapi.NullableDistributedPolicyDefaultsAndRulesV2) ( //nolint:unused
	*policy_openapi.SupersetDefaults, //worker
	*policy_openapi.SupersetDefaults, error) { //master

	if !distributed.IsSet() {
		return nil, nil, fmt.Errorf("failed to get distributed policy for sync")
	}

	distributedPolicy := distributed.Get()

	var workerDefaults policy_openapi.SupersetDefaults

	var masterDefaults *policy_openapi.SupersetDefaults

	errCopy := _copyDefaultsOrRules(&workerDefaults, distributedPolicy.GetDefaults().Worker.Get())
	if errCopy != nil {
		return nil, nil, fmt.Errorf("failed to load worker defaults of policy: %v", errCopy)
	}

	if distributedPolicy.GetRules().Master.IsSet() || distributedPolicy.GetDefaults().Master.IsSet() {
		masterDefaults = &policy_openapi.SupersetDefaults{}

		errCopy = _copyDefaultsOrRules(&masterDefaults, distributedPolicy.GetDefaults().Master.Get())
		if errCopy != nil {
			return nil, nil, fmt.Errorf("failed to load master defaults of policy: %v", errCopy)
		}
	}

	return &workerDefaults, masterDefaults, nil
}

func (p *PoliciesHandler) initStatus(backendPolicyCounter int64) workloadcommon.PolicyStatus {
	return workloadcommon.PolicyStatus{
		SyncStatus: p.getCommonSyncStatus(backendPolicyCounter),
	}
}

func (p *PoliciesHandler) initObjectMeta(backendPolicyMeta *policy_openapi.PolicyCreationFields,
	backendPolicyNamespace string, backendPolicyCounter int64) metav1.ObjectMeta {
	// Adding the status here because it isn't filled in from the backend.
	// We're aiming to minimize multiple API calls by combining the update and update status into a single call.
	objectMeta := metav1.ObjectMeta{
		Name: backendPolicyMeta.Name,
		Labels: map[string]string{
			constants.PolicyLabel: "true",
		},
		Namespace: backendPolicyNamespace,
		Annotations: map[string]string{
			constants.AnnotationsBackendPolicyCounter: fmt.Sprintf("%d", backendPolicyCounter),
		},
	}

	p.addPolicyLabels(&objectMeta, backendPolicyMeta)

	return objectMeta
}

func (p *PoliciesHandler) getCommonSyncStatus(policySyncCounter int64) workloadcommon.PolicySyncStatus {
	return workloadcommon.PolicySyncStatus{
		LastSyncRevision:          policySyncCounter,
		LastAttemptedSyncRevision: policySyncCounter,
		LastSyncTimestamp:         metav1.NewTime(time.Now()),
	}
}

func (p *PoliciesHandler) addPolicyLabels(objectMeta *metav1.ObjectMeta, backendPolicyMeta *policy_openapi.PolicyCreationFields) {
	switch backendPolicyMeta.Scope {
	case policy_openapi.SCOPETYPE_TENANT:
		objectMeta.GetLabels()[constants.TenantScopeLabel] = "true"
	case policy_openapi.SCOPETYPE_CLUSTER:
		objectMeta.GetLabels()[constants.ClusterWideResourceLabel] = "true"
	case policy_openapi.SCOPETYPE_DEPARTMENT:
		departmentIdPtr := backendPolicyMeta.DepartmentId.Get()
		if departmentIdPtr != nil {
			objectMeta.GetLabels()[constants.DepartmentScopeLabel] = *departmentIdPtr
		} else {
			log.Warn().Msgf("Had an error while getting backend policy scope id, name: %s, department id: %s", objectMeta.GetName(), *departmentIdPtr)
		}
	case policy_openapi.SCOPETYPE_PROJECT:
		projectIdPtr := backendPolicyMeta.ProjectId.Get()
		if projectIdPtr != nil {
			objectMeta.GetLabels()[constants.ProjectScopeLabel] = strconv.Itoa(int(*projectIdPtr))
		} else {
			log.Warn().Msgf("Had an error while getting backend policy scope id, name: %s, project id: %d", objectMeta.GetName(), *projectIdPtr)
		}
	}
}

func (p *PoliciesHandler) getBackendPolicies(prevHighestUpdate int64) (*policy_openapi.PolicySyncResponse, error) {
	var backendPolicy *policy_openapi.PolicySyncResponse
	var res *http.Response
	var err error

	syncReport, _ := syncReportPerPolicyType.Load(p.policyType)
	var theReport *policy_openapi.ClusterSyncReport = nil
	if syncReport.numPolicies > 0 {
		theReport = &syncReport.theReport
	}

	switch p.policyType {
	case common.PolicyTypeInteractive:
		backendPolicy, res, err = p.backendClient.PoliciesClient.SyncWorkspacePolicies(p.backendClient.Context(), config.Config().ClusterUuid, prevHighestUpdate, theReport)
		err = p.backendClient.HandleHttpResponse(res, err, "get backend workspace policy")
	case common.PolicyTypeTraining:
		backendPolicy, res, err = p.backendClient.PoliciesClient.SyncTrainingPolicies(p.backendClient.Context(), config.Config().ClusterUuid, prevHighestUpdate, theReport)
		err = p.backendClient.HandleHttpResponse(res, err, "get backend training policy")
	case common.PolicyTypeDistributed:
		backendPolicy, res, err = p.backendClient.PoliciesClient.SyncDistributedPolicies(p.backendClient.Context(), config.Config().ClusterUuid, prevHighestUpdate, theReport)
		err = p.backendClient.HandleHttpResponse(res, err, "get backend distributed policy")
	case common.PolicyTypeInference:
		backendPolicy, res, err = p.backendClient.PoliciesClient.SyncInferencePolicies(p.backendClient.Context(), config.Config().ClusterUuid, prevHighestUpdate, theReport)
		err = p.backendClient.HandleHttpResponse(res, err, "get backend inference policy")
	}

	if err != nil {
		log.Error().Msgf("Had an error while getting backend policy: %s", err.Error())
		return nil, err
	}

	return backendPolicy, nil
}

func NewPoliciesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader, policyType string) (ResourceHandler, error) {

	var object client.Object
	var objectInspector PoliciesObjectInspector

	switch policyType {
	case common.PolicyTypeInteractive:
		object = &ipv1.InteractivePolicy{}
		objectInspector = PoliciesObjectInspector(true)

	case common.PolicyTypeTraining:
		object = &tpv1.TrainingPolicy{}
		objectInspector = PoliciesObjectInspector(true)

	case common.PolicyTypeDistributed:
		object = &dpv1.DistributedPolicy{}
		objectInspector = PoliciesObjectInspector(true)

	case common.PolicyTypeInference:
		object = &infpv1.InferencePolicy{}
		objectInspector = PoliciesObjectInspector(true)

	default:
		return nil, fmt.Errorf("unsupported policy type: %s", policyType)
	}

	handlingOptions := HandlingOptions{deletePropagationOrphan: false}
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient, object, objectInspector, handlingOptions)
	if err != nil {
		return nil, err
	}

	agentConfigData, _ := configmaps.GetAgentConfig(cachedClient)
	switch policyType {
	case common.PolicyTypeInteractive:
		return &PoliciesHandler{
			CommonHandler: commonHandler,
			policyType:    policyType,
			syncCounter:   agentConfigData.InteractivePoliciesSyncCounter,
			metricLabel:   "interactive_policy",
		}, nil
	case common.PolicyTypeTraining:
		return &PoliciesHandler{
			CommonHandler: commonHandler,
			policyType:    policyType,
			syncCounter:   agentConfigData.TrainingPoliciesSyncCounter,
			metricLabel:   "training_policy",
		}, nil
	case common.PolicyTypeDistributed:
		return &PoliciesHandler{
			CommonHandler: commonHandler,
			policyType:    policyType,
			syncCounter:   agentConfigData.DistributedPoliciesSyncCounter,
			metricLabel:   "distributed_policy",
		}, nil
	case common.PolicyTypeInference:
		return &PoliciesHandler{
			CommonHandler: commonHandler,
			policyType:    policyType,
			syncCounter:   agentConfigData.InferencePoliciesSyncCounter,
			metricLabel:   "inference_policy",
		}, nil
	}

	return nil, fmt.Errorf("unsupported policy type: %s", policyType)
}

func NewInteractivePoliciesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewPoliciesHandler(runaiBackendClient, runaiClient, cachedClient, common.PolicyTypeInteractive)
}

func NewTrainingPoliciesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewPoliciesHandler(runaiBackendClient, runaiClient, cachedClient, common.PolicyTypeTraining)
}

func NewDistributedPoliciesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewPoliciesHandler(runaiBackendClient, runaiClient, cachedClient, common.PolicyTypeDistributed)
}

func NewInferencePoliciesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	return NewPoliciesHandler(runaiBackendClient, runaiClient, cachedClient, common.PolicyTypeInference)
}

func _copyDefaultsOrRules(dest interface{}, src interface{}) error { //nolint:unused
	asBytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(asBytes, dest)
	if err != nil {
		return err
	}
	return nil
}
