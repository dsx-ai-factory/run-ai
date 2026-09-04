# AIApplication

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** | The name of the application. The name must be unique within the organization and can only contain lowercase alphanumeric characters and hyphens. It must start with a letter. | [optional] 
**Source** | Pointer to **string** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**DesiredPhase** | Pointer to **NullableString** | The desired phase of the application. | [optional] 
**Phase** | Pointer to **string** | The actual status of the release in the cluster. | [optional] 
**ProjectId** | Pointer to **string** | The id of the project. | [optional] 
**ProjectName** | Pointer to **string** | The name of the project | [optional] 
**DepartmentId** | Pointer to **string** | The id of the department. | [optional] 
**DepartmentName** | Pointer to **string** | The name of the department | [optional] 
**Resources** | Pointer to [**[]AIApplicationResource**](AIApplicationResource.md) |  | [optional] 
**AllocatedResources** | Pointer to [**NullableAIApplicationAllocatedResources**](AIApplicationAllocatedResources.md) |  | [optional] 
**RequestedResources** | Pointer to [**NullableAIApplicationRequestResources**](AIApplicationRequestResources.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** | The user who created the application. | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedBy** | Pointer to **string** | The user who created the application. | [optional] 
**UpdateCount** | Pointer to **int64** | The sync update counter for this application. | [optional] 
**ActionStatus** | Pointer to **NullableString** | The status of the recent action preformed on the application. | [optional] 
**ActionMessage** | Pointer to **NullableString** | The message of the recent action preformed on the application. | [optional] 
**Endpoints** | Pointer to [**[]AiApplicationEndpoint**](AiApplicationEndpoint.md) | List of network endpoints through which the workloads associated with the AI application is exposed. | [optional] 
**ResourceConnections** | Pointer to [**[]AiApplicationResourceConnection**](AiApplicationResourceConnection.md) | Links between the application&#39;s workloads and the referenced resources (PersistentVolumeClaims, Secrets, ConfigMaps). Each entry connects one workload to one resource, so a resource shared by several workloads appears once per workload. | [optional] 
**Spec** | Pointer to [**AIApplicationSpec**](AIApplicationSpec.md) |  | [optional] 

## Methods

### NewAIApplication

`func NewAIApplication() *AIApplication`

NewAIApplication instantiates a new AIApplication object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationWithDefaults

`func NewAIApplicationWithDefaults() *AIApplication`

NewAIApplicationWithDefaults instantiates a new AIApplication object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AIApplication) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AIApplication) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AIApplication) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *AIApplication) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *AIApplication) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AIApplication) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AIApplication) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *AIApplication) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSource

`func (o *AIApplication) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *AIApplication) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *AIApplication) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *AIApplication) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetType

`func (o *AIApplication) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AIApplication) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AIApplication) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *AIApplication) HasType() bool`

HasType returns a boolean if a field has been set.

### GetDesiredPhase

`func (o *AIApplication) GetDesiredPhase() string`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *AIApplication) GetDesiredPhaseOk() (*string, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *AIApplication) SetDesiredPhase(v string)`

SetDesiredPhase sets DesiredPhase field to given value.

### HasDesiredPhase

`func (o *AIApplication) HasDesiredPhase() bool`

HasDesiredPhase returns a boolean if a field has been set.

### SetDesiredPhaseNil

`func (o *AIApplication) SetDesiredPhaseNil(b bool)`

 SetDesiredPhaseNil sets the value for DesiredPhase to be an explicit nil

### UnsetDesiredPhase
`func (o *AIApplication) UnsetDesiredPhase()`

UnsetDesiredPhase ensures that no value is present for DesiredPhase, not even an explicit nil
### GetPhase

`func (o *AIApplication) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *AIApplication) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *AIApplication) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *AIApplication) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetProjectId

`func (o *AIApplication) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AIApplication) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AIApplication) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *AIApplication) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetProjectName

`func (o *AIApplication) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *AIApplication) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *AIApplication) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *AIApplication) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetDepartmentId

`func (o *AIApplication) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *AIApplication) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *AIApplication) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *AIApplication) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetDepartmentName

`func (o *AIApplication) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *AIApplication) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *AIApplication) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *AIApplication) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetResources

`func (o *AIApplication) GetResources() []AIApplicationResource`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *AIApplication) GetResourcesOk() (*[]AIApplicationResource, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *AIApplication) SetResources(v []AIApplicationResource)`

SetResources sets Resources field to given value.

### HasResources

`func (o *AIApplication) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetAllocatedResources

`func (o *AIApplication) GetAllocatedResources() AIApplicationAllocatedResources`

GetAllocatedResources returns the AllocatedResources field if non-nil, zero value otherwise.

### GetAllocatedResourcesOk

`func (o *AIApplication) GetAllocatedResourcesOk() (*AIApplicationAllocatedResources, bool)`

GetAllocatedResourcesOk returns a tuple with the AllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResources

`func (o *AIApplication) SetAllocatedResources(v AIApplicationAllocatedResources)`

SetAllocatedResources sets AllocatedResources field to given value.

### HasAllocatedResources

`func (o *AIApplication) HasAllocatedResources() bool`

HasAllocatedResources returns a boolean if a field has been set.

### SetAllocatedResourcesNil

`func (o *AIApplication) SetAllocatedResourcesNil(b bool)`

 SetAllocatedResourcesNil sets the value for AllocatedResources to be an explicit nil

### UnsetAllocatedResources
`func (o *AIApplication) UnsetAllocatedResources()`

UnsetAllocatedResources ensures that no value is present for AllocatedResources, not even an explicit nil
### GetRequestedResources

`func (o *AIApplication) GetRequestedResources() AIApplicationRequestResources`

GetRequestedResources returns the RequestedResources field if non-nil, zero value otherwise.

### GetRequestedResourcesOk

`func (o *AIApplication) GetRequestedResourcesOk() (*AIApplicationRequestResources, bool)`

GetRequestedResourcesOk returns a tuple with the RequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedResources

`func (o *AIApplication) SetRequestedResources(v AIApplicationRequestResources)`

SetRequestedResources sets RequestedResources field to given value.

### HasRequestedResources

`func (o *AIApplication) HasRequestedResources() bool`

HasRequestedResources returns a boolean if a field has been set.

### SetRequestedResourcesNil

`func (o *AIApplication) SetRequestedResourcesNil(b bool)`

 SetRequestedResourcesNil sets the value for RequestedResources to be an explicit nil

### UnsetRequestedResources
`func (o *AIApplication) UnsetRequestedResources()`

UnsetRequestedResources ensures that no value is present for RequestedResources, not even an explicit nil
### GetCreatedAt

`func (o *AIApplication) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AIApplication) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AIApplication) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *AIApplication) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *AIApplication) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AIApplication) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AIApplication) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *AIApplication) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *AIApplication) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AIApplication) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AIApplication) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *AIApplication) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *AIApplication) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *AIApplication) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *AIApplication) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *AIApplication) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetUpdateCount

`func (o *AIApplication) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *AIApplication) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *AIApplication) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.

### HasUpdateCount

`func (o *AIApplication) HasUpdateCount() bool`

HasUpdateCount returns a boolean if a field has been set.

### GetActionStatus

`func (o *AIApplication) GetActionStatus() string`

GetActionStatus returns the ActionStatus field if non-nil, zero value otherwise.

### GetActionStatusOk

`func (o *AIApplication) GetActionStatusOk() (*string, bool)`

GetActionStatusOk returns a tuple with the ActionStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionStatus

`func (o *AIApplication) SetActionStatus(v string)`

SetActionStatus sets ActionStatus field to given value.

### HasActionStatus

`func (o *AIApplication) HasActionStatus() bool`

HasActionStatus returns a boolean if a field has been set.

### SetActionStatusNil

`func (o *AIApplication) SetActionStatusNil(b bool)`

 SetActionStatusNil sets the value for ActionStatus to be an explicit nil

### UnsetActionStatus
`func (o *AIApplication) UnsetActionStatus()`

UnsetActionStatus ensures that no value is present for ActionStatus, not even an explicit nil
### GetActionMessage

`func (o *AIApplication) GetActionMessage() string`

GetActionMessage returns the ActionMessage field if non-nil, zero value otherwise.

### GetActionMessageOk

`func (o *AIApplication) GetActionMessageOk() (*string, bool)`

GetActionMessageOk returns a tuple with the ActionMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionMessage

`func (o *AIApplication) SetActionMessage(v string)`

SetActionMessage sets ActionMessage field to given value.

### HasActionMessage

`func (o *AIApplication) HasActionMessage() bool`

HasActionMessage returns a boolean if a field has been set.

### SetActionMessageNil

`func (o *AIApplication) SetActionMessageNil(b bool)`

 SetActionMessageNil sets the value for ActionMessage to be an explicit nil

### UnsetActionMessage
`func (o *AIApplication) UnsetActionMessage()`

UnsetActionMessage ensures that no value is present for ActionMessage, not even an explicit nil
### GetEndpoints

`func (o *AIApplication) GetEndpoints() []AiApplicationEndpoint`

GetEndpoints returns the Endpoints field if non-nil, zero value otherwise.

### GetEndpointsOk

`func (o *AIApplication) GetEndpointsOk() (*[]AiApplicationEndpoint, bool)`

GetEndpointsOk returns a tuple with the Endpoints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEndpoints

`func (o *AIApplication) SetEndpoints(v []AiApplicationEndpoint)`

SetEndpoints sets Endpoints field to given value.

### HasEndpoints

`func (o *AIApplication) HasEndpoints() bool`

HasEndpoints returns a boolean if a field has been set.

### GetResourceConnections

`func (o *AIApplication) GetResourceConnections() []AiApplicationResourceConnection`

GetResourceConnections returns the ResourceConnections field if non-nil, zero value otherwise.

### GetResourceConnectionsOk

`func (o *AIApplication) GetResourceConnectionsOk() (*[]AiApplicationResourceConnection, bool)`

GetResourceConnectionsOk returns a tuple with the ResourceConnections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceConnections

`func (o *AIApplication) SetResourceConnections(v []AiApplicationResourceConnection)`

SetResourceConnections sets ResourceConnections field to given value.

### HasResourceConnections

`func (o *AIApplication) HasResourceConnections() bool`

HasResourceConnections returns a boolean if a field has been set.

### GetSpec

`func (o *AIApplication) GetSpec() AIApplicationSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *AIApplication) GetSpecOk() (*AIApplicationSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *AIApplication) SetSpec(v AIApplicationSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *AIApplication) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


