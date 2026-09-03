# WorkloadSyncFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Name** | **string** |  | 
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
**AdditionalFields** | Pointer to **string** |  | [optional] 
**Priority** | Pointer to **NullableInt32** |  | [optional] 
**FullySaasWorkload** | Pointer to **bool** |  | [optional] 

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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


