# AIApplicationAuth

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**UserCredential** | Pointer to **NullableBool** | Indicates whether the asset is a user credential owned by the caller. When omitted or false, the asset is treated as a shared credential. | [optional]
**AssetId** | Pointer to **string** | The ID of the asset to use for authentication. | [optional] 

## Methods

### NewAIApplicationAuth

`func NewAIApplicationAuth() *AIApplicationAuth`

NewAIApplicationAuth instantiates a new AIApplicationAuth object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationAuthWithDefaults

`func NewAIApplicationAuthWithDefaults() *AIApplicationAuth`

NewAIApplicationAuthWithDefaults instantiates a new AIApplicationAuth object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUserCredential

`func (o *AIApplicationAuth) GetUserCredential() bool`

GetUserCredential returns the UserCredential field if non-nil, zero value otherwise.

### GetUserCredentialOk

`func (o *AIApplicationAuth) GetUserCredentialOk() (*bool, bool)`

GetUserCredentialOk returns a tuple with the UserCredential field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserCredential

`func (o *AIApplicationAuth) SetUserCredential(v bool)`

SetUserCredential sets UserCredential field to given value.

### HasUserCredential

`func (o *AIApplicationAuth) HasUserCredential() bool`

HasUserCredential returns a boolean if a field has been set.

### SetUserCredentialNil

`func (o *AIApplicationAuth) SetUserCredentialNil(b bool)`

 SetUserCredentialNil sets the value for UserCredential to be an explicit nil

### UnsetUserCredential
`func (o *AIApplicationAuth) UnsetUserCredential()`

UnsetUserCredential ensures that no value is present for UserCredential, not even an explicit nil
### GetAssetId

`func (o *AIApplicationAuth) GetAssetId() string`

GetAssetId returns the AssetId field if non-nil, zero value otherwise.

### GetAssetIdOk

`func (o *AIApplicationAuth) GetAssetIdOk() (*string, bool)`

GetAssetIdOk returns a tuple with the AssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetId

`func (o *AIApplicationAuth) SetAssetId(v string)`

SetAssetId sets AssetId field to given value.

### HasAssetId

`func (o *AIApplicationAuth) HasAssetId() bool`

HasAssetId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


