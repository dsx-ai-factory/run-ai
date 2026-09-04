# WorkloadV2Update

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | A unique ID of the workload. | 
**Name** | **string** | The name of the workload. | 
**CreatedBy** | **string** | Immutable subject ID of the user who created the workload | 
**ProjectId** | Pointer to **string** | The id of the project. | [optional] 
**Gvk** | [**GVK**](GVK.md) |  | 
**UpdateCount** | **int64** | Number of updates made to the workload | [default to 0]
**DesiredPhase** | [**DesiredPhase**](DesiredPhase.md) |  | 
**Configuration** | Pointer to [**WorkloadV2UpdateConfiguration**](WorkloadV2UpdateConfiguration.md) |  | [optional] 
**WorkloadManifest** | Pointer to **map[string]interface{}** | The full Kubernetes YAML manifest for the workload (only present in some updates) | [optional] 

## Methods

### NewWorkloadV2Update

`func NewWorkloadV2Update(id string, name string, createdBy string, gvk GVK, updateCount int64, desiredPhase DesiredPhase, ) *WorkloadV2Update`

NewWorkloadV2Update instantiates a new WorkloadV2Update object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2UpdateWithDefaults

`func NewWorkloadV2UpdateWithDefaults() *WorkloadV2Update`

NewWorkloadV2UpdateWithDefaults instantiates a new WorkloadV2Update object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadV2Update) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadV2Update) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadV2Update) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *WorkloadV2Update) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadV2Update) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadV2Update) SetName(v string)`

SetName sets Name field to given value.


### GetCreatedBy

`func (o *WorkloadV2Update) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *WorkloadV2Update) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *WorkloadV2Update) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetProjectId

`func (o *WorkloadV2Update) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadV2Update) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadV2Update) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *WorkloadV2Update) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetGvk

`func (o *WorkloadV2Update) GetGvk() GVK`

GetGvk returns the Gvk field if non-nil, zero value otherwise.

### GetGvkOk

`func (o *WorkloadV2Update) GetGvkOk() (*GVK, bool)`

GetGvkOk returns a tuple with the Gvk field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGvk

`func (o *WorkloadV2Update) SetGvk(v GVK)`

SetGvk sets Gvk field to given value.


### GetUpdateCount

`func (o *WorkloadV2Update) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *WorkloadV2Update) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *WorkloadV2Update) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.


### GetDesiredPhase

`func (o *WorkloadV2Update) GetDesiredPhase() DesiredPhase`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *WorkloadV2Update) GetDesiredPhaseOk() (*DesiredPhase, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *WorkloadV2Update) SetDesiredPhase(v DesiredPhase)`

SetDesiredPhase sets DesiredPhase field to given value.


### GetConfiguration

`func (o *WorkloadV2Update) GetConfiguration() WorkloadV2UpdateConfiguration`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *WorkloadV2Update) GetConfigurationOk() (*WorkloadV2UpdateConfiguration, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *WorkloadV2Update) SetConfiguration(v WorkloadV2UpdateConfiguration)`

SetConfiguration sets Configuration field to given value.

### HasConfiguration

`func (o *WorkloadV2Update) HasConfiguration() bool`

HasConfiguration returns a boolean if a field has been set.

### GetWorkloadManifest

`func (o *WorkloadV2Update) GetWorkloadManifest() map[string]interface{}`

GetWorkloadManifest returns the WorkloadManifest field if non-nil, zero value otherwise.

### GetWorkloadManifestOk

`func (o *WorkloadV2Update) GetWorkloadManifestOk() (*map[string]interface{}, bool)`

GetWorkloadManifestOk returns a tuple with the WorkloadManifest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadManifest

`func (o *WorkloadV2Update) SetWorkloadManifest(v map[string]interface{})`

SetWorkloadManifest sets WorkloadManifest field to given value.

### HasWorkloadManifest

`func (o *WorkloadV2Update) HasWorkloadManifest() bool`

HasWorkloadManifest returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


