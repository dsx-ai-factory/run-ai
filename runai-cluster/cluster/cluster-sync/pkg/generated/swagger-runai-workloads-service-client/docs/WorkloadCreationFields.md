# WorkloadCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** |  | 
**Name** | **string** |  | 
**Id** | **string** |  | 
**SubmittedBy** | Pointer to **string** |  | [optional] 
**ClusterId** | **string** |  | 
**ProjectId** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**RequestedResources** | Pointer to [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | [optional] 
**RequestedPods** | Pointer to [**RequestedPods**](RequestedPods.md) |  | [optional] 
**RequestedNodePools** | Pointer to **[]string** |  | [optional] 
**Datasources** | Pointer to [**[]Datasource**](Datasource.md) |  | [optional] 
**Environments** | Pointer to [**[]Environment**](Environment.md) |  | [optional] 
**ExternalConnections** | Pointer to [**[]Connection**](Connection.md) |  | [optional] 
**DistributedFramework** | Pointer to **string** |  | [optional] 
**Images** | Pointer to **[]string** |  | [optional] 

## Methods

### NewWorkloadCreationFields

`func NewWorkloadCreationFields(type_ string, name string, id string, clusterId string, projectId string, createdAt time.Time, ) *WorkloadCreationFields`

NewWorkloadCreationFields instantiates a new WorkloadCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadCreationFieldsWithDefaults

`func NewWorkloadCreationFieldsWithDefaults() *WorkloadCreationFields`

NewWorkloadCreationFieldsWithDefaults instantiates a new WorkloadCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *WorkloadCreationFields) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *WorkloadCreationFields) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *WorkloadCreationFields) SetType(v string)`

SetType sets Type field to given value.


### GetName

`func (o *WorkloadCreationFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadCreationFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadCreationFields) SetName(v string)`

SetName sets Name field to given value.


### GetId

`func (o *WorkloadCreationFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadCreationFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadCreationFields) SetId(v string)`

SetId sets Id field to given value.


### GetSubmittedBy

`func (o *WorkloadCreationFields) GetSubmittedBy() string`

GetSubmittedBy returns the SubmittedBy field if non-nil, zero value otherwise.

### GetSubmittedByOk

`func (o *WorkloadCreationFields) GetSubmittedByOk() (*string, bool)`

GetSubmittedByOk returns a tuple with the SubmittedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubmittedBy

`func (o *WorkloadCreationFields) SetSubmittedBy(v string)`

SetSubmittedBy sets SubmittedBy field to given value.

### HasSubmittedBy

`func (o *WorkloadCreationFields) HasSubmittedBy() bool`

HasSubmittedBy returns a boolean if a field has been set.

### GetClusterId

`func (o *WorkloadCreationFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadCreationFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadCreationFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *WorkloadCreationFields) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadCreationFields) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadCreationFields) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetCreatedAt

`func (o *WorkloadCreationFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadCreationFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadCreationFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetRequestedResources

`func (o *WorkloadCreationFields) GetRequestedResources() WorkloadRequestResources`

GetRequestedResources returns the RequestedResources field if non-nil, zero value otherwise.

### GetRequestedResourcesOk

`func (o *WorkloadCreationFields) GetRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetRequestedResourcesOk returns a tuple with the RequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedResources

`func (o *WorkloadCreationFields) SetRequestedResources(v WorkloadRequestResources)`

SetRequestedResources sets RequestedResources field to given value.

### HasRequestedResources

`func (o *WorkloadCreationFields) HasRequestedResources() bool`

HasRequestedResources returns a boolean if a field has been set.

### SetRequestedResourcesNil

`func (o *WorkloadCreationFields) SetRequestedResourcesNil(b bool)`

 SetRequestedResourcesNil sets the value for RequestedResources to be an explicit nil

### UnsetRequestedResources
`func (o *WorkloadCreationFields) UnsetRequestedResources()`

UnsetRequestedResources ensures that no value is present for RequestedResources, not even an explicit nil
### GetRequestedPods

`func (o *WorkloadCreationFields) GetRequestedPods() RequestedPods`

GetRequestedPods returns the RequestedPods field if non-nil, zero value otherwise.

### GetRequestedPodsOk

`func (o *WorkloadCreationFields) GetRequestedPodsOk() (*RequestedPods, bool)`

GetRequestedPodsOk returns a tuple with the RequestedPods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedPods

`func (o *WorkloadCreationFields) SetRequestedPods(v RequestedPods)`

SetRequestedPods sets RequestedPods field to given value.

### HasRequestedPods

`func (o *WorkloadCreationFields) HasRequestedPods() bool`

HasRequestedPods returns a boolean if a field has been set.

### GetRequestedNodePools

`func (o *WorkloadCreationFields) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *WorkloadCreationFields) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *WorkloadCreationFields) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *WorkloadCreationFields) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetDatasources

`func (o *WorkloadCreationFields) GetDatasources() []Datasource`

GetDatasources returns the Datasources field if non-nil, zero value otherwise.

### GetDatasourcesOk

`func (o *WorkloadCreationFields) GetDatasourcesOk() (*[]Datasource, bool)`

GetDatasourcesOk returns a tuple with the Datasources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasources

`func (o *WorkloadCreationFields) SetDatasources(v []Datasource)`

SetDatasources sets Datasources field to given value.

### HasDatasources

`func (o *WorkloadCreationFields) HasDatasources() bool`

HasDatasources returns a boolean if a field has been set.

### GetEnvironments

`func (o *WorkloadCreationFields) GetEnvironments() []Environment`

GetEnvironments returns the Environments field if non-nil, zero value otherwise.

### GetEnvironmentsOk

`func (o *WorkloadCreationFields) GetEnvironmentsOk() (*[]Environment, bool)`

GetEnvironmentsOk returns a tuple with the Environments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironments

`func (o *WorkloadCreationFields) SetEnvironments(v []Environment)`

SetEnvironments sets Environments field to given value.

### HasEnvironments

`func (o *WorkloadCreationFields) HasEnvironments() bool`

HasEnvironments returns a boolean if a field has been set.

### GetExternalConnections

`func (o *WorkloadCreationFields) GetExternalConnections() []Connection`

GetExternalConnections returns the ExternalConnections field if non-nil, zero value otherwise.

### GetExternalConnectionsOk

`func (o *WorkloadCreationFields) GetExternalConnectionsOk() (*[]Connection, bool)`

GetExternalConnectionsOk returns a tuple with the ExternalConnections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalConnections

`func (o *WorkloadCreationFields) SetExternalConnections(v []Connection)`

SetExternalConnections sets ExternalConnections field to given value.

### HasExternalConnections

`func (o *WorkloadCreationFields) HasExternalConnections() bool`

HasExternalConnections returns a boolean if a field has been set.

### GetDistributedFramework

`func (o *WorkloadCreationFields) GetDistributedFramework() string`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *WorkloadCreationFields) GetDistributedFrameworkOk() (*string, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *WorkloadCreationFields) SetDistributedFramework(v string)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *WorkloadCreationFields) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### GetImages

`func (o *WorkloadCreationFields) GetImages() []string`

GetImages returns the Images field if non-nil, zero value otherwise.

### GetImagesOk

`func (o *WorkloadCreationFields) GetImagesOk() (*[]string, bool)`

GetImagesOk returns a tuple with the Images field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImages

`func (o *WorkloadCreationFields) SetImages(v []string)`

SetImages sets Images field to given value.

### HasImages

`func (o *WorkloadCreationFields) HasImages() bool`

HasImages returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


