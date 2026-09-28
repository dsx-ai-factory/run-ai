# WorkloadSyncFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Name** | **string** |  | 
**Group** | Pointer to **string** |  | [optional] 
**Version** | Pointer to **string** |  | [optional] 
**ExternalWorkloadName** | Pointer to **string** |  | [optional] 
**Id** | **string** |  | 
**SubmittedBy** | Pointer to **string** |  | [optional] 
**PriorityClassName** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectName** | **string** |  | 
**ProjectId** | **string** |  | 
**DepartmentName** | **string** |  | 
**DepartmentId** | **string** |  | 
**Namespace** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**RequestedResources** | Pointer to [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | [optional] 
**ActionsSupport** | Pointer to [**ActionsSupport**](ActionsSupport.md) |  | [optional] 
**Phase** | [**Phase**](Phase.md) |  | 
**Source** | [**Source**](Source.md) |  | 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**PhaseMessage** | Pointer to **string** |  | [optional] 
**K8sPhase** | **string** |  | 
**RequestedPods** | Pointer to [**RequestedPods**](RequestedPods.md) |  | [optional] 
**RequestedNodePools** | Pointer to **[]string** |  | [optional] 
**Images** | Pointer to **[]string** |  | [optional] 
**DistributedFramework** | Pointer to **string** |  | [optional] 
**ChildrenIds** | Pointer to [**[]WorkloadChildrenIdsInner**](WorkloadChildrenIdsInner.md) |  | [optional] 
**Urls** | Pointer to **[]string** |  | [optional] 
**NetworkStatus** | Pointer to [**WorkloadNetworkStatus**](WorkloadNetworkStatus.md) |  | [optional] 
**AdditionalFields** | Pointer to **string** |  | [optional] 
**Priority** | Pointer to **NullableInt32** |  | [optional] 
**FullySaasWorkload** | Pointer to **bool** |  | [optional] 
**EnvironmentVariables** | Pointer to **map[string]string** |  | [optional] 
**Command** | Pointer to **string** |  | [optional] 
**Arguments** | Pointer to **string** |  | [optional] 
**SchedulingConditions** | Pointer to [**[]SchedulingCondition**](SchedulingCondition.md) |  | [optional] 
**Category** | Pointer to **string** |  | [optional] 
**GuaranteedRuntimeEndsAt** | Pointer to **NullableTime** | A timestamp indicating when the workload will reach its minimum guaranteed runtime, as defined by minGuaranteedRuntime. Until this time, the workload is considered non-preemptible and cannot be interrupted by higher-priority workloads. | [optional] 
**ApplicationId** | Pointer to **NullableString** |  | [optional] 
**ApplicationName** | Pointer to **NullableString** |  | [optional] 
**Topology** | Pointer to [**NullableWorkloadSyncFieldsTopology**](WorkloadSyncFieldsTopology.md) |  | [optional] 
**Preemptibility** | Pointer to **NullableString** |  | [optional] 
**StructureElements** | Pointer to [**[]WorkloadElement**](WorkloadElement.md) | Flat list of intermediate child elements. May be truncated if too many. | [optional] 
**StructureElementsTotalCount** | Pointer to **NullableInt32** | Total number of structure elements before truncation. Null if no structure exists. | [optional] 
**Endpoints** | Pointer to [**[]EndpointSync**](EndpointSync.md) |  | [optional] 
**NodeAffinities** | Pointer to [**[]WorkloadNodeAffinity**](WorkloadNodeAffinity.md) | The node affinities of a workload, which is an array of node affinity rules for different specs (e.g., worker, leader, master). Each item in the array specifies the node affinity for a particular spec. | [optional] 
**ResourceRefs** | Pointer to [**[]ResourceRef**](ResourceRef.md) | Existing resources (PVCs, Secrets, ConfigMaps) referenced by the workload&#39;s pod template. Surfaced at the application level as resourceConnections. | [optional] 

## Methods

### NewWorkloadSyncFields

`func NewWorkloadSyncFields(type_ string, name string, id string, priorityClassName string, clusterId string, projectName string, projectId string, departmentName string, departmentId string, namespace string, createdAt time.Time, phase Phase, source Source, k8sPhase string, ) *WorkloadSyncFields`

NewWorkloadSyncFields instantiates a new WorkloadSyncFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadSyncFieldsWithDefaults

`func NewWorkloadSyncFieldsWithDefaults() *WorkloadSyncFields`

NewWorkloadSyncFieldsWithDefaults instantiates a new WorkloadSyncFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *WorkloadSyncFields) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *WorkloadSyncFields) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *WorkloadSyncFields) SetType(v string)`

SetType sets Type field to given value.


### GetName

`func (o *WorkloadSyncFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadSyncFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadSyncFields) SetName(v string)`

SetName sets Name field to given value.


### GetGroup

`func (o *WorkloadSyncFields) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *WorkloadSyncFields) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *WorkloadSyncFields) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *WorkloadSyncFields) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

### GetVersion

`func (o *WorkloadSyncFields) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadSyncFields) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadSyncFields) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadSyncFields) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetExternalWorkloadName

`func (o *WorkloadSyncFields) GetExternalWorkloadName() string`

GetExternalWorkloadName returns the ExternalWorkloadName field if non-nil, zero value otherwise.

### GetExternalWorkloadNameOk

`func (o *WorkloadSyncFields) GetExternalWorkloadNameOk() (*string, bool)`

GetExternalWorkloadNameOk returns a tuple with the ExternalWorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalWorkloadName

`func (o *WorkloadSyncFields) SetExternalWorkloadName(v string)`

SetExternalWorkloadName sets ExternalWorkloadName field to given value.

### HasExternalWorkloadName

`func (o *WorkloadSyncFields) HasExternalWorkloadName() bool`

HasExternalWorkloadName returns a boolean if a field has been set.

### GetId

`func (o *WorkloadSyncFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadSyncFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadSyncFields) SetId(v string)`

SetId sets Id field to given value.


### GetSubmittedBy

`func (o *WorkloadSyncFields) GetSubmittedBy() string`

GetSubmittedBy returns the SubmittedBy field if non-nil, zero value otherwise.

### GetSubmittedByOk

`func (o *WorkloadSyncFields) GetSubmittedByOk() (*string, bool)`

GetSubmittedByOk returns a tuple with the SubmittedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubmittedBy

`func (o *WorkloadSyncFields) SetSubmittedBy(v string)`

SetSubmittedBy sets SubmittedBy field to given value.

### HasSubmittedBy

`func (o *WorkloadSyncFields) HasSubmittedBy() bool`

HasSubmittedBy returns a boolean if a field has been set.

### GetPriorityClassName

`func (o *WorkloadSyncFields) GetPriorityClassName() string`

GetPriorityClassName returns the PriorityClassName field if non-nil, zero value otherwise.

### GetPriorityClassNameOk

`func (o *WorkloadSyncFields) GetPriorityClassNameOk() (*string, bool)`

GetPriorityClassNameOk returns a tuple with the PriorityClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClassName

`func (o *WorkloadSyncFields) SetPriorityClassName(v string)`

SetPriorityClassName sets PriorityClassName field to given value.


### GetClusterId

`func (o *WorkloadSyncFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadSyncFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadSyncFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectName

`func (o *WorkloadSyncFields) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *WorkloadSyncFields) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *WorkloadSyncFields) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetProjectId

`func (o *WorkloadSyncFields) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadSyncFields) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadSyncFields) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetDepartmentName

`func (o *WorkloadSyncFields) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *WorkloadSyncFields) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *WorkloadSyncFields) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.


### GetDepartmentId

`func (o *WorkloadSyncFields) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *WorkloadSyncFields) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *WorkloadSyncFields) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.


### GetNamespace

`func (o *WorkloadSyncFields) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *WorkloadSyncFields) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *WorkloadSyncFields) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetCreatedAt

`func (o *WorkloadSyncFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadSyncFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadSyncFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetRequestedResources

`func (o *WorkloadSyncFields) GetRequestedResources() WorkloadRequestResources`

GetRequestedResources returns the RequestedResources field if non-nil, zero value otherwise.

### GetRequestedResourcesOk

`func (o *WorkloadSyncFields) GetRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetRequestedResourcesOk returns a tuple with the RequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedResources

`func (o *WorkloadSyncFields) SetRequestedResources(v WorkloadRequestResources)`

SetRequestedResources sets RequestedResources field to given value.

### HasRequestedResources

`func (o *WorkloadSyncFields) HasRequestedResources() bool`

HasRequestedResources returns a boolean if a field has been set.

### SetRequestedResourcesNil

`func (o *WorkloadSyncFields) SetRequestedResourcesNil(b bool)`

 SetRequestedResourcesNil sets the value for RequestedResources to be an explicit nil

### UnsetRequestedResources
`func (o *WorkloadSyncFields) UnsetRequestedResources()`

UnsetRequestedResources ensures that no value is present for RequestedResources, not even an explicit nil
### GetActionsSupport

`func (o *WorkloadSyncFields) GetActionsSupport() ActionsSupport`

GetActionsSupport returns the ActionsSupport field if non-nil, zero value otherwise.

### GetActionsSupportOk

`func (o *WorkloadSyncFields) GetActionsSupportOk() (*ActionsSupport, bool)`

GetActionsSupportOk returns a tuple with the ActionsSupport field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionsSupport

`func (o *WorkloadSyncFields) SetActionsSupport(v ActionsSupport)`

SetActionsSupport sets ActionsSupport field to given value.

### HasActionsSupport

`func (o *WorkloadSyncFields) HasActionsSupport() bool`

HasActionsSupport returns a boolean if a field has been set.

### GetPhase

`func (o *WorkloadSyncFields) GetPhase() Phase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *WorkloadSyncFields) GetPhaseOk() (*Phase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *WorkloadSyncFields) SetPhase(v Phase)`

SetPhase sets Phase field to given value.


### GetSource

`func (o *WorkloadSyncFields) GetSource() Source`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *WorkloadSyncFields) GetSourceOk() (*Source, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *WorkloadSyncFields) SetSource(v Source)`

SetSource sets Source field to given value.


### GetConditions

`func (o *WorkloadSyncFields) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *WorkloadSyncFields) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *WorkloadSyncFields) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *WorkloadSyncFields) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetPhaseMessage

`func (o *WorkloadSyncFields) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *WorkloadSyncFields) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *WorkloadSyncFields) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *WorkloadSyncFields) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetK8sPhase

`func (o *WorkloadSyncFields) GetK8sPhase() string`

GetK8sPhase returns the K8sPhase field if non-nil, zero value otherwise.

### GetK8sPhaseOk

`func (o *WorkloadSyncFields) GetK8sPhaseOk() (*string, bool)`

GetK8sPhaseOk returns a tuple with the K8sPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhase

`func (o *WorkloadSyncFields) SetK8sPhase(v string)`

SetK8sPhase sets K8sPhase field to given value.


### GetRequestedPods

`func (o *WorkloadSyncFields) GetRequestedPods() RequestedPods`

GetRequestedPods returns the RequestedPods field if non-nil, zero value otherwise.

### GetRequestedPodsOk

`func (o *WorkloadSyncFields) GetRequestedPodsOk() (*RequestedPods, bool)`

GetRequestedPodsOk returns a tuple with the RequestedPods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedPods

`func (o *WorkloadSyncFields) SetRequestedPods(v RequestedPods)`

SetRequestedPods sets RequestedPods field to given value.

### HasRequestedPods

`func (o *WorkloadSyncFields) HasRequestedPods() bool`

HasRequestedPods returns a boolean if a field has been set.

### GetRequestedNodePools

`func (o *WorkloadSyncFields) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *WorkloadSyncFields) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *WorkloadSyncFields) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *WorkloadSyncFields) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetImages

`func (o *WorkloadSyncFields) GetImages() []string`

GetImages returns the Images field if non-nil, zero value otherwise.

### GetImagesOk

`func (o *WorkloadSyncFields) GetImagesOk() (*[]string, bool)`

GetImagesOk returns a tuple with the Images field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImages

`func (o *WorkloadSyncFields) SetImages(v []string)`

SetImages sets Images field to given value.

### HasImages

`func (o *WorkloadSyncFields) HasImages() bool`

HasImages returns a boolean if a field has been set.

### GetDistributedFramework

`func (o *WorkloadSyncFields) GetDistributedFramework() string`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *WorkloadSyncFields) GetDistributedFrameworkOk() (*string, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *WorkloadSyncFields) SetDistributedFramework(v string)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *WorkloadSyncFields) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### GetChildrenIds

`func (o *WorkloadSyncFields) GetChildrenIds() []WorkloadChildrenIdsInner`

GetChildrenIds returns the ChildrenIds field if non-nil, zero value otherwise.

### GetChildrenIdsOk

`func (o *WorkloadSyncFields) GetChildrenIdsOk() (*[]WorkloadChildrenIdsInner, bool)`

GetChildrenIdsOk returns a tuple with the ChildrenIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildrenIds

`func (o *WorkloadSyncFields) SetChildrenIds(v []WorkloadChildrenIdsInner)`

SetChildrenIds sets ChildrenIds field to given value.

### HasChildrenIds

`func (o *WorkloadSyncFields) HasChildrenIds() bool`

HasChildrenIds returns a boolean if a field has been set.

### GetUrls

`func (o *WorkloadSyncFields) GetUrls() []string`

GetUrls returns the Urls field if non-nil, zero value otherwise.

### GetUrlsOk

`func (o *WorkloadSyncFields) GetUrlsOk() (*[]string, bool)`

GetUrlsOk returns a tuple with the Urls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrls

`func (o *WorkloadSyncFields) SetUrls(v []string)`

SetUrls sets Urls field to given value.

### HasUrls

`func (o *WorkloadSyncFields) HasUrls() bool`

HasUrls returns a boolean if a field has been set.

### GetNetworkStatus

`func (o *WorkloadSyncFields) GetNetworkStatus() WorkloadNetworkStatus`

GetNetworkStatus returns the NetworkStatus field if non-nil, zero value otherwise.

### GetNetworkStatusOk

`func (o *WorkloadSyncFields) GetNetworkStatusOk() (*WorkloadNetworkStatus, bool)`

GetNetworkStatusOk returns a tuple with the NetworkStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkStatus

`func (o *WorkloadSyncFields) SetNetworkStatus(v WorkloadNetworkStatus)`

SetNetworkStatus sets NetworkStatus field to given value.

### HasNetworkStatus

`func (o *WorkloadSyncFields) HasNetworkStatus() bool`

HasNetworkStatus returns a boolean if a field has been set.

### GetAdditionalFields

`func (o *WorkloadSyncFields) GetAdditionalFields() string`

GetAdditionalFields returns the AdditionalFields field if non-nil, zero value otherwise.

### GetAdditionalFieldsOk

`func (o *WorkloadSyncFields) GetAdditionalFieldsOk() (*string, bool)`

GetAdditionalFieldsOk returns a tuple with the AdditionalFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalFields

`func (o *WorkloadSyncFields) SetAdditionalFields(v string)`

SetAdditionalFields sets AdditionalFields field to given value.

### HasAdditionalFields

`func (o *WorkloadSyncFields) HasAdditionalFields() bool`

HasAdditionalFields returns a boolean if a field has been set.

### GetPriority

`func (o *WorkloadSyncFields) GetPriority() int32`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *WorkloadSyncFields) GetPriorityOk() (*int32, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *WorkloadSyncFields) SetPriority(v int32)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *WorkloadSyncFields) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### SetPriorityNil

`func (o *WorkloadSyncFields) SetPriorityNil(b bool)`

 SetPriorityNil sets the value for Priority to be an explicit nil

### UnsetPriority
`func (o *WorkloadSyncFields) UnsetPriority()`

UnsetPriority ensures that no value is present for Priority, not even an explicit nil
### GetFullySaasWorkload

`func (o *WorkloadSyncFields) GetFullySaasWorkload() bool`

GetFullySaasWorkload returns the FullySaasWorkload field if non-nil, zero value otherwise.

### GetFullySaasWorkloadOk

`func (o *WorkloadSyncFields) GetFullySaasWorkloadOk() (*bool, bool)`

GetFullySaasWorkloadOk returns a tuple with the FullySaasWorkload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullySaasWorkload

`func (o *WorkloadSyncFields) SetFullySaasWorkload(v bool)`

SetFullySaasWorkload sets FullySaasWorkload field to given value.

### HasFullySaasWorkload

`func (o *WorkloadSyncFields) HasFullySaasWorkload() bool`

HasFullySaasWorkload returns a boolean if a field has been set.

### GetEnvironmentVariables

`func (o *WorkloadSyncFields) GetEnvironmentVariables() map[string]string`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *WorkloadSyncFields) GetEnvironmentVariablesOk() (*map[string]string, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *WorkloadSyncFields) SetEnvironmentVariables(v map[string]string)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *WorkloadSyncFields) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### GetCommand

`func (o *WorkloadSyncFields) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *WorkloadSyncFields) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *WorkloadSyncFields) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *WorkloadSyncFields) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetArguments

`func (o *WorkloadSyncFields) GetArguments() string`

GetArguments returns the Arguments field if non-nil, zero value otherwise.

### GetArgumentsOk

`func (o *WorkloadSyncFields) GetArgumentsOk() (*string, bool)`

GetArgumentsOk returns a tuple with the Arguments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArguments

`func (o *WorkloadSyncFields) SetArguments(v string)`

SetArguments sets Arguments field to given value.

### HasArguments

`func (o *WorkloadSyncFields) HasArguments() bool`

HasArguments returns a boolean if a field has been set.

### GetSchedulingConditions

`func (o *WorkloadSyncFields) GetSchedulingConditions() []SchedulingCondition`

GetSchedulingConditions returns the SchedulingConditions field if non-nil, zero value otherwise.

### GetSchedulingConditionsOk

`func (o *WorkloadSyncFields) GetSchedulingConditionsOk() (*[]SchedulingCondition, bool)`

GetSchedulingConditionsOk returns a tuple with the SchedulingConditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingConditions

`func (o *WorkloadSyncFields) SetSchedulingConditions(v []SchedulingCondition)`

SetSchedulingConditions sets SchedulingConditions field to given value.

### HasSchedulingConditions

`func (o *WorkloadSyncFields) HasSchedulingConditions() bool`

HasSchedulingConditions returns a boolean if a field has been set.

### SetSchedulingConditionsNil

`func (o *WorkloadSyncFields) SetSchedulingConditionsNil(b bool)`

 SetSchedulingConditionsNil sets the value for SchedulingConditions to be an explicit nil

### UnsetSchedulingConditions
`func (o *WorkloadSyncFields) UnsetSchedulingConditions()`

UnsetSchedulingConditions ensures that no value is present for SchedulingConditions, not even an explicit nil
### GetCategory

`func (o *WorkloadSyncFields) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkloadSyncFields) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkloadSyncFields) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkloadSyncFields) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### GetGuaranteedRuntimeEndsAt

`func (o *WorkloadSyncFields) GetGuaranteedRuntimeEndsAt() time.Time`

GetGuaranteedRuntimeEndsAt returns the GuaranteedRuntimeEndsAt field if non-nil, zero value otherwise.

### GetGuaranteedRuntimeEndsAtOk

`func (o *WorkloadSyncFields) GetGuaranteedRuntimeEndsAtOk() (*time.Time, bool)`

GetGuaranteedRuntimeEndsAtOk returns a tuple with the GuaranteedRuntimeEndsAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuaranteedRuntimeEndsAt

`func (o *WorkloadSyncFields) SetGuaranteedRuntimeEndsAt(v time.Time)`

SetGuaranteedRuntimeEndsAt sets GuaranteedRuntimeEndsAt field to given value.

### HasGuaranteedRuntimeEndsAt

`func (o *WorkloadSyncFields) HasGuaranteedRuntimeEndsAt() bool`

HasGuaranteedRuntimeEndsAt returns a boolean if a field has been set.

### SetGuaranteedRuntimeEndsAtNil

`func (o *WorkloadSyncFields) SetGuaranteedRuntimeEndsAtNil(b bool)`

 SetGuaranteedRuntimeEndsAtNil sets the value for GuaranteedRuntimeEndsAt to be an explicit nil

### UnsetGuaranteedRuntimeEndsAt
`func (o *WorkloadSyncFields) UnsetGuaranteedRuntimeEndsAt()`

UnsetGuaranteedRuntimeEndsAt ensures that no value is present for GuaranteedRuntimeEndsAt, not even an explicit nil
### GetApplicationId

`func (o *WorkloadSyncFields) GetApplicationId() string`

GetApplicationId returns the ApplicationId field if non-nil, zero value otherwise.

### GetApplicationIdOk

`func (o *WorkloadSyncFields) GetApplicationIdOk() (*string, bool)`

GetApplicationIdOk returns a tuple with the ApplicationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplicationId

`func (o *WorkloadSyncFields) SetApplicationId(v string)`

SetApplicationId sets ApplicationId field to given value.

### HasApplicationId

`func (o *WorkloadSyncFields) HasApplicationId() bool`

HasApplicationId returns a boolean if a field has been set.

### SetApplicationIdNil

`func (o *WorkloadSyncFields) SetApplicationIdNil(b bool)`

 SetApplicationIdNil sets the value for ApplicationId to be an explicit nil

### UnsetApplicationId
`func (o *WorkloadSyncFields) UnsetApplicationId()`

UnsetApplicationId ensures that no value is present for ApplicationId, not even an explicit nil
### GetApplicationName

`func (o *WorkloadSyncFields) GetApplicationName() string`

GetApplicationName returns the ApplicationName field if non-nil, zero value otherwise.

### GetApplicationNameOk

`func (o *WorkloadSyncFields) GetApplicationNameOk() (*string, bool)`

GetApplicationNameOk returns a tuple with the ApplicationName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplicationName

`func (o *WorkloadSyncFields) SetApplicationName(v string)`

SetApplicationName sets ApplicationName field to given value.

### HasApplicationName

`func (o *WorkloadSyncFields) HasApplicationName() bool`

HasApplicationName returns a boolean if a field has been set.

### SetApplicationNameNil

`func (o *WorkloadSyncFields) SetApplicationNameNil(b bool)`

 SetApplicationNameNil sets the value for ApplicationName to be an explicit nil

### UnsetApplicationName
`func (o *WorkloadSyncFields) UnsetApplicationName()`

UnsetApplicationName ensures that no value is present for ApplicationName, not even an explicit nil
### GetTopology

`func (o *WorkloadSyncFields) GetTopology() WorkloadSyncFieldsTopology`

GetTopology returns the Topology field if non-nil, zero value otherwise.

### GetTopologyOk

`func (o *WorkloadSyncFields) GetTopologyOk() (*WorkloadSyncFieldsTopology, bool)`

GetTopologyOk returns a tuple with the Topology field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopology

`func (o *WorkloadSyncFields) SetTopology(v WorkloadSyncFieldsTopology)`

SetTopology sets Topology field to given value.

### HasTopology

`func (o *WorkloadSyncFields) HasTopology() bool`

HasTopology returns a boolean if a field has been set.

### SetTopologyNil

`func (o *WorkloadSyncFields) SetTopologyNil(b bool)`

 SetTopologyNil sets the value for Topology to be an explicit nil

### UnsetTopology
`func (o *WorkloadSyncFields) UnsetTopology()`

UnsetTopology ensures that no value is present for Topology, not even an explicit nil
### GetPreemptibility

`func (o *WorkloadSyncFields) GetPreemptibility() string`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadSyncFields) GetPreemptibilityOk() (*string, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadSyncFields) SetPreemptibility(v string)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadSyncFields) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkloadSyncFields) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkloadSyncFields) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetStructureElements

`func (o *WorkloadSyncFields) GetStructureElements() []WorkloadElement`

GetStructureElements returns the StructureElements field if non-nil, zero value otherwise.

### GetStructureElementsOk

`func (o *WorkloadSyncFields) GetStructureElementsOk() (*[]WorkloadElement, bool)`

GetStructureElementsOk returns a tuple with the StructureElements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStructureElements

`func (o *WorkloadSyncFields) SetStructureElements(v []WorkloadElement)`

SetStructureElements sets StructureElements field to given value.

### HasStructureElements

`func (o *WorkloadSyncFields) HasStructureElements() bool`

HasStructureElements returns a boolean if a field has been set.

### SetStructureElementsNil

`func (o *WorkloadSyncFields) SetStructureElementsNil(b bool)`

 SetStructureElementsNil sets the value for StructureElements to be an explicit nil

### UnsetStructureElements
`func (o *WorkloadSyncFields) UnsetStructureElements()`

UnsetStructureElements ensures that no value is present for StructureElements, not even an explicit nil
### GetStructureElementsTotalCount

`func (o *WorkloadSyncFields) GetStructureElementsTotalCount() int32`

GetStructureElementsTotalCount returns the StructureElementsTotalCount field if non-nil, zero value otherwise.

### GetStructureElementsTotalCountOk

`func (o *WorkloadSyncFields) GetStructureElementsTotalCountOk() (*int32, bool)`

GetStructureElementsTotalCountOk returns a tuple with the StructureElementsTotalCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStructureElementsTotalCount

`func (o *WorkloadSyncFields) SetStructureElementsTotalCount(v int32)`

SetStructureElementsTotalCount sets StructureElementsTotalCount field to given value.

### HasStructureElementsTotalCount

`func (o *WorkloadSyncFields) HasStructureElementsTotalCount() bool`

HasStructureElementsTotalCount returns a boolean if a field has been set.

### SetStructureElementsTotalCountNil

`func (o *WorkloadSyncFields) SetStructureElementsTotalCountNil(b bool)`

 SetStructureElementsTotalCountNil sets the value for StructureElementsTotalCount to be an explicit nil

### UnsetStructureElementsTotalCount
`func (o *WorkloadSyncFields) UnsetStructureElementsTotalCount()`

UnsetStructureElementsTotalCount ensures that no value is present for StructureElementsTotalCount, not even an explicit nil
### GetEndpoints

`func (o *WorkloadSyncFields) GetEndpoints() []EndpointSync`

GetEndpoints returns the Endpoints field if non-nil, zero value otherwise.

### GetEndpointsOk

`func (o *WorkloadSyncFields) GetEndpointsOk() (*[]EndpointSync, bool)`

GetEndpointsOk returns a tuple with the Endpoints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEndpoints

`func (o *WorkloadSyncFields) SetEndpoints(v []EndpointSync)`

SetEndpoints sets Endpoints field to given value.

### HasEndpoints

`func (o *WorkloadSyncFields) HasEndpoints() bool`

HasEndpoints returns a boolean if a field has been set.

### GetNodeAffinities

`func (o *WorkloadSyncFields) GetNodeAffinities() []WorkloadNodeAffinity`

GetNodeAffinities returns the NodeAffinities field if non-nil, zero value otherwise.

### GetNodeAffinitiesOk

`func (o *WorkloadSyncFields) GetNodeAffinitiesOk() (*[]WorkloadNodeAffinity, bool)`

GetNodeAffinitiesOk returns a tuple with the NodeAffinities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinities

`func (o *WorkloadSyncFields) SetNodeAffinities(v []WorkloadNodeAffinity)`

SetNodeAffinities sets NodeAffinities field to given value.

### HasNodeAffinities

`func (o *WorkloadSyncFields) HasNodeAffinities() bool`

HasNodeAffinities returns a boolean if a field has been set.

### GetResourceRefs

`func (o *WorkloadSyncFields) GetResourceRefs() []ResourceRef`

GetResourceRefs returns the ResourceRefs field if non-nil, zero value otherwise.

### GetResourceRefsOk

`func (o *WorkloadSyncFields) GetResourceRefsOk() (*[]ResourceRef, bool)`

GetResourceRefsOk returns a tuple with the ResourceRefs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceRefs

`func (o *WorkloadSyncFields) SetResourceRefs(v []ResourceRef)`

SetResourceRefs sets ResourceRefs field to given value.

### HasResourceRefs

`func (o *WorkloadSyncFields) HasResourceRefs() bool`

HasResourceRefs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


