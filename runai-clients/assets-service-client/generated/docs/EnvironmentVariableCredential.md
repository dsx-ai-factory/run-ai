# EnvironmentVariableCredential

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AssetId** | Pointer to **NullableString** | Unique identifier of the asset. (mandatory) | [optional] 
**Key** | Pointer to **NullableString** | The key in the secret resource. (mandatory) | [optional] 

## Methods

### NewEnvironmentVariableCredential

`func NewEnvironmentVariableCredential() *EnvironmentVariableCredential`

NewEnvironmentVariableCredential instantiates a new EnvironmentVariableCredential object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentVariableCredentialWithDefaults

`func NewEnvironmentVariableCredentialWithDefaults() *EnvironmentVariableCredential`

NewEnvironmentVariableCredentialWithDefaults instantiates a new EnvironmentVariableCredential object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssetId

`func (o *EnvironmentVariableCredential) GetAssetId() string`

GetAssetId returns the AssetId field if non-nil, zero value otherwise.

### GetAssetIdOk

`func (o *EnvironmentVariableCredential) GetAssetIdOk() (*string, bool)`

GetAssetIdOk returns a tuple with the AssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetId

`func (o *EnvironmentVariableCredential) SetAssetId(v string)`

SetAssetId sets AssetId field to given value.

### HasAssetId

`func (o *EnvironmentVariableCredential) HasAssetId() bool`

HasAssetId returns a boolean if a field has been set.

### SetAssetIdNil

`func (o *EnvironmentVariableCredential) SetAssetIdNil(b bool)`

 SetAssetIdNil sets the value for AssetId to be an explicit nil

### UnsetAssetId
`func (o *EnvironmentVariableCredential) UnsetAssetId()`

UnsetAssetId ensures that no value is present for AssetId, not even an explicit nil
### GetKey

`func (o *EnvironmentVariableCredential) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *EnvironmentVariableCredential) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *EnvironmentVariableCredential) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *EnvironmentVariableCredential) HasKey() bool`

HasKey returns a boolean if a field has been set.

### SetKeyNil

`func (o *EnvironmentVariableCredential) SetKeyNil(b bool)`

 SetKeyNil sets the value for Key to be an explicit nil

### UnsetKey
`func (o *EnvironmentVariableCredential) UnsetKey()`

UnsetKey ensures that no value is present for Key, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


