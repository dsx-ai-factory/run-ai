# DistributedCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**UseGivenNameAsPrefix** | Pointer to **bool** | When true, the requested name will be treated as a prefix. The final name of the workload will be composed of the name followed by a random set of characters. | [optional] [default to false]
**ProjectId** | **string** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 
**TemplateId** | Pointer to **NullableString** | The unique identifier of the template to use for submitting this workload. The combined values provided in the spec, template and assets will be used to create the workload. | [optional] 
**Assets** | Pointer to [**NullableAssetSubmissionIds**](AssetSubmissionIds.md) |  | [optional] 
**Spec** | Pointer to [**DistributedSpecSpec**](DistributedSpecSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**MasterSpec** | Pointer to [**NullableMasterSpec**](MasterSpec.md) |  | [optional] 
**MasterAssets** | Pointer to [**NullableAssetSubmissionIds**](AssetSubmissionIds.md) |  | [optional] 

## Methods

### NewDistributedCreationRequest

`func NewDistributedCreationRequest(name string, projectId string, clusterId string, ) *DistributedCreationRequest`

NewDistributedCreationRequest instantiates a new DistributedCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedCreationRequestWithDefaults

`func NewDistributedCreationRequestWithDefaults() *DistributedCreationRequest`

NewDistributedCreationRequestWithDefaults instantiates a new DistributedCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DistributedCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DistributedCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DistributedCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUseGivenNameAsPrefix

`func (o *DistributedCreationRequest) GetUseGivenNameAsPrefix() bool`

GetUseGivenNameAsPrefix returns the UseGivenNameAsPrefix field if non-nil, zero value otherwise.

### GetUseGivenNameAsPrefixOk

`func (o *DistributedCreationRequest) GetUseGivenNameAsPrefixOk() (*bool, bool)`

GetUseGivenNameAsPrefixOk returns a tuple with the UseGivenNameAsPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseGivenNameAsPrefix

`func (o *DistributedCreationRequest) SetUseGivenNameAsPrefix(v bool)`

SetUseGivenNameAsPrefix sets UseGivenNameAsPrefix field to given value.

### HasUseGivenNameAsPrefix

`func (o *DistributedCreationRequest) HasUseGivenNameAsPrefix() bool`

HasUseGivenNameAsPrefix returns a boolean if a field has been set.

### GetProjectId

`func (o *DistributedCreationRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *DistributedCreationRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *DistributedCreationRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *DistributedCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DistributedCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DistributedCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetTemplateId

`func (o *DistributedCreationRequest) GetTemplateId() string`

GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.

### GetTemplateIdOk

`func (o *DistributedCreationRequest) GetTemplateIdOk() (*string, bool)`

GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplateId

`func (o *DistributedCreationRequest) SetTemplateId(v string)`

SetTemplateId sets TemplateId field to given value.

### HasTemplateId

`func (o *DistributedCreationRequest) HasTemplateId() bool`

HasTemplateId returns a boolean if a field has been set.

### SetTemplateIdNil

`func (o *DistributedCreationRequest) SetTemplateIdNil(b bool)`

 SetTemplateIdNil sets the value for TemplateId to be an explicit nil

### UnsetTemplateId
`func (o *DistributedCreationRequest) UnsetTemplateId()`

UnsetTemplateId ensures that no value is present for TemplateId, not even an explicit nil
### GetAssets

`func (o *DistributedCreationRequest) GetAssets() AssetSubmissionIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *DistributedCreationRequest) GetAssetsOk() (*AssetSubmissionIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *DistributedCreationRequest) SetAssets(v AssetSubmissionIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *DistributedCreationRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### SetAssetsNil

`func (o *DistributedCreationRequest) SetAssetsNil(b bool)`

 SetAssetsNil sets the value for Assets to be an explicit nil

### UnsetAssets
`func (o *DistributedCreationRequest) UnsetAssets()`

UnsetAssets ensures that no value is present for Assets, not even an explicit nil
### GetSpec

`func (o *DistributedCreationRequest) GetSpec() DistributedSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedCreationRequest) GetSpecOk() (*DistributedSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedCreationRequest) SetSpec(v DistributedSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedCreationRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *DistributedCreationRequest) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *DistributedCreationRequest) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *DistributedCreationRequest) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *DistributedCreationRequest) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *DistributedCreationRequest) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *DistributedCreationRequest) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMasterSpec

`func (o *DistributedCreationRequest) GetMasterSpec() MasterSpec`

GetMasterSpec returns the MasterSpec field if non-nil, zero value otherwise.

### GetMasterSpecOk

`func (o *DistributedCreationRequest) GetMasterSpecOk() (*MasterSpec, bool)`

GetMasterSpecOk returns a tuple with the MasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpec

`func (o *DistributedCreationRequest) SetMasterSpec(v MasterSpec)`

SetMasterSpec sets MasterSpec field to given value.

### HasMasterSpec

`func (o *DistributedCreationRequest) HasMasterSpec() bool`

HasMasterSpec returns a boolean if a field has been set.

### SetMasterSpecNil

`func (o *DistributedCreationRequest) SetMasterSpecNil(b bool)`

 SetMasterSpecNil sets the value for MasterSpec to be an explicit nil

### UnsetMasterSpec
`func (o *DistributedCreationRequest) UnsetMasterSpec()`

UnsetMasterSpec ensures that no value is present for MasterSpec, not even an explicit nil
### GetMasterAssets

`func (o *DistributedCreationRequest) GetMasterAssets() AssetSubmissionIds`

GetMasterAssets returns the MasterAssets field if non-nil, zero value otherwise.

### GetMasterAssetsOk

`func (o *DistributedCreationRequest) GetMasterAssetsOk() (*AssetSubmissionIds, bool)`

GetMasterAssetsOk returns a tuple with the MasterAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterAssets

`func (o *DistributedCreationRequest) SetMasterAssets(v AssetSubmissionIds)`

SetMasterAssets sets MasterAssets field to given value.

### HasMasterAssets

`func (o *DistributedCreationRequest) HasMasterAssets() bool`

HasMasterAssets returns a boolean if a field has been set.

### SetMasterAssetsNil

`func (o *DistributedCreationRequest) SetMasterAssetsNil(b bool)`

 SetMasterAssetsNil sets the value for MasterAssets to be an explicit nil

### UnsetMasterAssets
`func (o *DistributedCreationRequest) UnsetMasterAssets()`

UnsetMasterAssets ensures that no value is present for MasterAssets, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


