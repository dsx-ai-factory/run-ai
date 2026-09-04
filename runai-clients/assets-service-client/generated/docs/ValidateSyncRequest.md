# ValidateSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** |  | 
**ProjectId** | **string** |  | 
**AssetKind** | [**AssetKind**](AssetKind.md) |  | 
**AssetNames** | **[]string** | List of the assets which was marked for deletion | 

## Methods

### NewValidateSyncRequest

`func NewValidateSyncRequest(clusterId string, projectId string, assetKind AssetKind, assetNames []string, ) *ValidateSyncRequest`

NewValidateSyncRequest instantiates a new ValidateSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewValidateSyncRequestWithDefaults

`func NewValidateSyncRequestWithDefaults() *ValidateSyncRequest`

NewValidateSyncRequestWithDefaults instantiates a new ValidateSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *ValidateSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ValidateSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ValidateSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *ValidateSyncRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *ValidateSyncRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *ValidateSyncRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetAssetKind

`func (o *ValidateSyncRequest) GetAssetKind() AssetKind`

GetAssetKind returns the AssetKind field if non-nil, zero value otherwise.

### GetAssetKindOk

`func (o *ValidateSyncRequest) GetAssetKindOk() (*AssetKind, bool)`

GetAssetKindOk returns a tuple with the AssetKind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetKind

`func (o *ValidateSyncRequest) SetAssetKind(v AssetKind)`

SetAssetKind sets AssetKind field to given value.


### GetAssetNames

`func (o *ValidateSyncRequest) GetAssetNames() []string`

GetAssetNames returns the AssetNames field if non-nil, zero value otherwise.

### GetAssetNamesOk

`func (o *ValidateSyncRequest) GetAssetNamesOk() (*[]string, bool)`

GetAssetNamesOk returns a tuple with the AssetNames field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetNames

`func (o *ValidateSyncRequest) SetAssetNames(v []string)`

SetAssetNames sets AssetNames field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


