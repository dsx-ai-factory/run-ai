# AssetFlatRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AssetId** | **string** |  | 
**AssetIds** | Pointer to **[]string** |  | [optional] 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectId** | **string** | The id of the project. | 
**WorkloadType** | [**WorkloadType**](WorkloadType.md) |  | 
**MasterSameAsWorker** | Pointer to **bool** |  | [optional] 

## Methods

### NewAssetFlatRequest

`func NewAssetFlatRequest(assetId string, clusterId string, projectId string, workloadType WorkloadType, ) *AssetFlatRequest`

NewAssetFlatRequest instantiates a new AssetFlatRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetFlatRequestWithDefaults

`func NewAssetFlatRequestWithDefaults() *AssetFlatRequest`

NewAssetFlatRequestWithDefaults instantiates a new AssetFlatRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssetId

`func (o *AssetFlatRequest) GetAssetId() string`

GetAssetId returns the AssetId field if non-nil, zero value otherwise.

### GetAssetIdOk

`func (o *AssetFlatRequest) GetAssetIdOk() (*string, bool)`

GetAssetIdOk returns a tuple with the AssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetId

`func (o *AssetFlatRequest) SetAssetId(v string)`

SetAssetId sets AssetId field to given value.


### GetAssetIds

`func (o *AssetFlatRequest) GetAssetIds() []string`

GetAssetIds returns the AssetIds field if non-nil, zero value otherwise.

### GetAssetIdsOk

`func (o *AssetFlatRequest) GetAssetIdsOk() (*[]string, bool)`

GetAssetIdsOk returns a tuple with the AssetIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetIds

`func (o *AssetFlatRequest) SetAssetIds(v []string)`

SetAssetIds sets AssetIds field to given value.

### HasAssetIds

`func (o *AssetFlatRequest) HasAssetIds() bool`

HasAssetIds returns a boolean if a field has been set.

### GetClusterId

`func (o *AssetFlatRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AssetFlatRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AssetFlatRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *AssetFlatRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AssetFlatRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AssetFlatRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetWorkloadType

`func (o *AssetFlatRequest) GetWorkloadType() WorkloadType`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *AssetFlatRequest) GetWorkloadTypeOk() (*WorkloadType, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *AssetFlatRequest) SetWorkloadType(v WorkloadType)`

SetWorkloadType sets WorkloadType field to given value.


### GetMasterSameAsWorker

`func (o *AssetFlatRequest) GetMasterSameAsWorker() bool`

GetMasterSameAsWorker returns the MasterSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSameAsWorkerOk

`func (o *AssetFlatRequest) GetMasterSameAsWorkerOk() (*bool, bool)`

GetMasterSameAsWorkerOk returns a tuple with the MasterSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSameAsWorker

`func (o *AssetFlatRequest) SetMasterSameAsWorker(v bool)`

SetMasterSameAsWorker sets MasterSameAsWorker field to given value.

### HasMasterSameAsWorker

`func (o *AssetFlatRequest) HasMasterSameAsWorker() bool`

HasMasterSameAsWorker returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


