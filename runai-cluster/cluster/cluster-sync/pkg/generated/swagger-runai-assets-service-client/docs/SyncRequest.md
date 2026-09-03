# SyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | The id of the cluster. If empty - tenant scoped assets. | [optional] 
**ProjectName** | Pointer to **string** | The name of the project in which the assets reside. If empty - tenant scoped assets. | [optional] 
**Assets** | Pointer to [**[]AssetNameAndChecksum**](AssetNameAndChecksum.md) | List of the assets which exist in the cluster | [optional] 

## Methods

### NewSyncRequest

`func NewSyncRequest() *SyncRequest`

NewSyncRequest instantiates a new SyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncRequestWithDefaults

`func NewSyncRequestWithDefaults() *SyncRequest`

NewSyncRequestWithDefaults instantiates a new SyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *SyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *SyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *SyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *SyncRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetProjectName

`func (o *SyncRequest) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *SyncRequest) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *SyncRequest) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *SyncRequest) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetAssets

`func (o *SyncRequest) GetAssets() []AssetNameAndChecksum`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *SyncRequest) GetAssetsOk() (*[]AssetNameAndChecksum, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *SyncRequest) SetAssets(v []AssetNameAndChecksum)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *SyncRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


