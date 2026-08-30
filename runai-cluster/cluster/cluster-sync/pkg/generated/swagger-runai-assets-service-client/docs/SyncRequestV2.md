# SyncRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | The id of the cluster. If empty - tenant scoped assets. | [optional] 
**Scope** | Pointer to [**Scope**](Scope.md) |  | [optional] 
**ScopeResourceId** | Pointer to **string** | the id of the scope resource, empty for tenant scope, the id of the department for department scope and the project name for project scope. | [optional] 
**Assets** | Pointer to [**[]AssetNameAndChecksum**](AssetNameAndChecksum.md) | List of the assets which exist in the cluster | [optional] 

## Methods

### NewSyncRequestV2

`func NewSyncRequestV2() *SyncRequestV2`

NewSyncRequestV2 instantiates a new SyncRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncRequestV2WithDefaults

`func NewSyncRequestV2WithDefaults() *SyncRequestV2`

NewSyncRequestV2WithDefaults instantiates a new SyncRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *SyncRequestV2) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *SyncRequestV2) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *SyncRequestV2) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *SyncRequestV2) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetScope

`func (o *SyncRequestV2) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *SyncRequestV2) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *SyncRequestV2) SetScope(v Scope)`

SetScope sets Scope field to given value.

### HasScope

`func (o *SyncRequestV2) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetScopeResourceId

`func (o *SyncRequestV2) GetScopeResourceId() string`

GetScopeResourceId returns the ScopeResourceId field if non-nil, zero value otherwise.

### GetScopeResourceIdOk

`func (o *SyncRequestV2) GetScopeResourceIdOk() (*string, bool)`

GetScopeResourceIdOk returns a tuple with the ScopeResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeResourceId

`func (o *SyncRequestV2) SetScopeResourceId(v string)`

SetScopeResourceId sets ScopeResourceId field to given value.

### HasScopeResourceId

`func (o *SyncRequestV2) HasScopeResourceId() bool`

HasScopeResourceId returns a boolean if a field has been set.

### GetAssets

`func (o *SyncRequestV2) GetAssets() []AssetNameAndChecksum`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *SyncRequestV2) GetAssetsOk() (*[]AssetNameAndChecksum, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *SyncRequestV2) SetAssets(v []AssetNameAndChecksum)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *SyncRequestV2) HasAssets() bool`

HasAssets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


