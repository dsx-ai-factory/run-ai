# AssetChangeNotification

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChangeType** | Pointer to [**AssetChangeType**](AssetChangeType.md) |  | [optional] 
**ScopeType** | Pointer to [**ScopeType**](ScopeType.md) |  | [optional] 
**ScopeId** | Pointer to **NullableString** | cluster, department or project id, depending on the scope type | [optional] 

## Methods

### NewAssetChangeNotification

`func NewAssetChangeNotification() *AssetChangeNotification`

NewAssetChangeNotification instantiates a new AssetChangeNotification object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetChangeNotificationWithDefaults

`func NewAssetChangeNotificationWithDefaults() *AssetChangeNotification`

NewAssetChangeNotificationWithDefaults instantiates a new AssetChangeNotification object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChangeType

`func (o *AssetChangeNotification) GetChangeType() AssetChangeType`

GetChangeType returns the ChangeType field if non-nil, zero value otherwise.

### GetChangeTypeOk

`func (o *AssetChangeNotification) GetChangeTypeOk() (*AssetChangeType, bool)`

GetChangeTypeOk returns a tuple with the ChangeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChangeType

`func (o *AssetChangeNotification) SetChangeType(v AssetChangeType)`

SetChangeType sets ChangeType field to given value.

### HasChangeType

`func (o *AssetChangeNotification) HasChangeType() bool`

HasChangeType returns a boolean if a field has been set.

### GetScopeType

`func (o *AssetChangeNotification) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AssetChangeNotification) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AssetChangeNotification) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.

### HasScopeType

`func (o *AssetChangeNotification) HasScopeType() bool`

HasScopeType returns a boolean if a field has been set.

### GetScopeId

`func (o *AssetChangeNotification) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AssetChangeNotification) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AssetChangeNotification) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *AssetChangeNotification) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.

### SetScopeIdNil

`func (o *AssetChangeNotification) SetScopeIdNil(b bool)`

 SetScopeIdNil sets the value for ScopeId to be an explicit nil

### UnsetScopeId
`func (o *AssetChangeNotification) UnsetScopeId()`

UnsetScopeId ensures that no value is present for ScopeId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


