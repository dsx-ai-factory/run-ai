# UnifiedWorkloadRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadType** | Pointer to [**WorkloadType**](WorkloadType.md) |  | [optional] 
**ProjectId** | Pointer to **string** | The id of the project. | [optional] 
**ClusterId** | Pointer to **NullableString** | The id of the cluster. | [optional] 
**AssetId** | Pointer to **NullableString** | Optional unique identifier of the asset. | [optional] 
**AssetIds** | Pointer to **[]string** |  | [optional] 
**DataVolumeIds** | Pointer to **[]string** |  | [optional] 
**WorkloadId** | Pointer to **NullableString** | A unique ID of the workload. | [optional] 
**TemplateId** | Pointer to **NullableString** | A unique ID of the template. | [optional] 

## Methods

### NewUnifiedWorkloadRequest

`func NewUnifiedWorkloadRequest() *UnifiedWorkloadRequest`

NewUnifiedWorkloadRequest instantiates a new UnifiedWorkloadRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUnifiedWorkloadRequestWithDefaults

`func NewUnifiedWorkloadRequestWithDefaults() *UnifiedWorkloadRequest`

NewUnifiedWorkloadRequestWithDefaults instantiates a new UnifiedWorkloadRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadType

`func (o *UnifiedWorkloadRequest) GetWorkloadType() WorkloadType`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *UnifiedWorkloadRequest) GetWorkloadTypeOk() (*WorkloadType, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *UnifiedWorkloadRequest) SetWorkloadType(v WorkloadType)`

SetWorkloadType sets WorkloadType field to given value.

### HasWorkloadType

`func (o *UnifiedWorkloadRequest) HasWorkloadType() bool`

HasWorkloadType returns a boolean if a field has been set.

### GetProjectId

`func (o *UnifiedWorkloadRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *UnifiedWorkloadRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *UnifiedWorkloadRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *UnifiedWorkloadRequest) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetClusterId

`func (o *UnifiedWorkloadRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *UnifiedWorkloadRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *UnifiedWorkloadRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *UnifiedWorkloadRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *UnifiedWorkloadRequest) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *UnifiedWorkloadRequest) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetAssetId

`func (o *UnifiedWorkloadRequest) GetAssetId() string`

GetAssetId returns the AssetId field if non-nil, zero value otherwise.

### GetAssetIdOk

`func (o *UnifiedWorkloadRequest) GetAssetIdOk() (*string, bool)`

GetAssetIdOk returns a tuple with the AssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetId

`func (o *UnifiedWorkloadRequest) SetAssetId(v string)`

SetAssetId sets AssetId field to given value.

### HasAssetId

`func (o *UnifiedWorkloadRequest) HasAssetId() bool`

HasAssetId returns a boolean if a field has been set.

### SetAssetIdNil

`func (o *UnifiedWorkloadRequest) SetAssetIdNil(b bool)`

 SetAssetIdNil sets the value for AssetId to be an explicit nil

### UnsetAssetId
`func (o *UnifiedWorkloadRequest) UnsetAssetId()`

UnsetAssetId ensures that no value is present for AssetId, not even an explicit nil
### GetAssetIds

`func (o *UnifiedWorkloadRequest) GetAssetIds() []string`

GetAssetIds returns the AssetIds field if non-nil, zero value otherwise.

### GetAssetIdsOk

`func (o *UnifiedWorkloadRequest) GetAssetIdsOk() (*[]string, bool)`

GetAssetIdsOk returns a tuple with the AssetIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetIds

`func (o *UnifiedWorkloadRequest) SetAssetIds(v []string)`

SetAssetIds sets AssetIds field to given value.

### HasAssetIds

`func (o *UnifiedWorkloadRequest) HasAssetIds() bool`

HasAssetIds returns a boolean if a field has been set.

### GetDataVolumeIds

`func (o *UnifiedWorkloadRequest) GetDataVolumeIds() []string`

GetDataVolumeIds returns the DataVolumeIds field if non-nil, zero value otherwise.

### GetDataVolumeIdsOk

`func (o *UnifiedWorkloadRequest) GetDataVolumeIdsOk() (*[]string, bool)`

GetDataVolumeIdsOk returns a tuple with the DataVolumeIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolumeIds

`func (o *UnifiedWorkloadRequest) SetDataVolumeIds(v []string)`

SetDataVolumeIds sets DataVolumeIds field to given value.

### HasDataVolumeIds

`func (o *UnifiedWorkloadRequest) HasDataVolumeIds() bool`

HasDataVolumeIds returns a boolean if a field has been set.

### GetWorkloadId

`func (o *UnifiedWorkloadRequest) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *UnifiedWorkloadRequest) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *UnifiedWorkloadRequest) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.

### HasWorkloadId

`func (o *UnifiedWorkloadRequest) HasWorkloadId() bool`

HasWorkloadId returns a boolean if a field has been set.

### SetWorkloadIdNil

`func (o *UnifiedWorkloadRequest) SetWorkloadIdNil(b bool)`

 SetWorkloadIdNil sets the value for WorkloadId to be an explicit nil

### UnsetWorkloadId
`func (o *UnifiedWorkloadRequest) UnsetWorkloadId()`

UnsetWorkloadId ensures that no value is present for WorkloadId, not even an explicit nil
### GetTemplateId

`func (o *UnifiedWorkloadRequest) GetTemplateId() string`

GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.

### GetTemplateIdOk

`func (o *UnifiedWorkloadRequest) GetTemplateIdOk() (*string, bool)`

GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplateId

`func (o *UnifiedWorkloadRequest) SetTemplateId(v string)`

SetTemplateId sets TemplateId field to given value.

### HasTemplateId

`func (o *UnifiedWorkloadRequest) HasTemplateId() bool`

HasTemplateId returns a boolean if a field has been set.

### SetTemplateIdNil

`func (o *UnifiedWorkloadRequest) SetTemplateIdNil(b bool)`

 SetTemplateIdNil sets the value for TemplateId to be an explicit nil

### UnsetTemplateId
`func (o *UnifiedWorkloadRequest) UnsetTemplateId()`

UnsetTemplateId ensures that no value is present for TemplateId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


