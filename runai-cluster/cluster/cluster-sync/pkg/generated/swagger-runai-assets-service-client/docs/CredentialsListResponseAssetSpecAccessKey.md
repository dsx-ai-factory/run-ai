# CredentialsListResponseAssetSpecAccessKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | optional name of existing secret in the cluster from which the access key id and secret should be taken. If omitted, you will have to provide those credentials in the accessKeyId and secretAccessKey fields. The provided credentials are encrypted into the control plane database and cloned to a kubernetes secret in the cluster. | [optional] 
**AccessKeyId** | Pointer to **NullableString** | The access key id of the S3-compatible bucket, required only when not using existing secret. | [optional] 
**SecretAccessKey** | Pointer to **NullableString** | The secret access key of the S3-compatible bucket, required only when not using existing secret. | [optional] 
**KeyId** | **int32** | id of the encryption key which has been used to encrypt an asset&#39;s data. | 

## Methods

### NewCredentialsListResponseAssetSpecAccessKey

`func NewCredentialsListResponseAssetSpecAccessKey(keyId int32, ) *CredentialsListResponseAssetSpecAccessKey`

NewCredentialsListResponseAssetSpecAccessKey instantiates a new CredentialsListResponseAssetSpecAccessKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCredentialsListResponseAssetSpecAccessKeyWithDefaults

`func NewCredentialsListResponseAssetSpecAccessKeyWithDefaults() *CredentialsListResponseAssetSpecAccessKey`

NewCredentialsListResponseAssetSpecAccessKeyWithDefaults instantiates a new CredentialsListResponseAssetSpecAccessKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *CredentialsListResponseAssetSpecAccessKey) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *CredentialsListResponseAssetSpecAccessKey) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *CredentialsListResponseAssetSpecAccessKey) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *CredentialsListResponseAssetSpecAccessKey) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *CredentialsListResponseAssetSpecAccessKey) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *CredentialsListResponseAssetSpecAccessKey) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetAccessKeyId

`func (o *CredentialsListResponseAssetSpecAccessKey) GetAccessKeyId() string`

GetAccessKeyId returns the AccessKeyId field if non-nil, zero value otherwise.

### GetAccessKeyIdOk

`func (o *CredentialsListResponseAssetSpecAccessKey) GetAccessKeyIdOk() (*string, bool)`

GetAccessKeyIdOk returns a tuple with the AccessKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeyId

`func (o *CredentialsListResponseAssetSpecAccessKey) SetAccessKeyId(v string)`

SetAccessKeyId sets AccessKeyId field to given value.

### HasAccessKeyId

`func (o *CredentialsListResponseAssetSpecAccessKey) HasAccessKeyId() bool`

HasAccessKeyId returns a boolean if a field has been set.

### SetAccessKeyIdNil

`func (o *CredentialsListResponseAssetSpecAccessKey) SetAccessKeyIdNil(b bool)`

 SetAccessKeyIdNil sets the value for AccessKeyId to be an explicit nil

### UnsetAccessKeyId
`func (o *CredentialsListResponseAssetSpecAccessKey) UnsetAccessKeyId()`

UnsetAccessKeyId ensures that no value is present for AccessKeyId, not even an explicit nil
### GetSecretAccessKey

`func (o *CredentialsListResponseAssetSpecAccessKey) GetSecretAccessKey() string`

GetSecretAccessKey returns the SecretAccessKey field if non-nil, zero value otherwise.

### GetSecretAccessKeyOk

`func (o *CredentialsListResponseAssetSpecAccessKey) GetSecretAccessKeyOk() (*string, bool)`

GetSecretAccessKeyOk returns a tuple with the SecretAccessKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretAccessKey

`func (o *CredentialsListResponseAssetSpecAccessKey) SetSecretAccessKey(v string)`

SetSecretAccessKey sets SecretAccessKey field to given value.

### HasSecretAccessKey

`func (o *CredentialsListResponseAssetSpecAccessKey) HasSecretAccessKey() bool`

HasSecretAccessKey returns a boolean if a field has been set.

### SetSecretAccessKeyNil

`func (o *CredentialsListResponseAssetSpecAccessKey) SetSecretAccessKeyNil(b bool)`

 SetSecretAccessKeyNil sets the value for SecretAccessKey to be an explicit nil

### UnsetSecretAccessKey
`func (o *CredentialsListResponseAssetSpecAccessKey) UnsetSecretAccessKey()`

UnsetSecretAccessKey ensures that no value is present for SecretAccessKey, not even an explicit nil
### GetKeyId

`func (o *CredentialsListResponseAssetSpecAccessKey) GetKeyId() int32`

GetKeyId returns the KeyId field if non-nil, zero value otherwise.

### GetKeyIdOk

`func (o *CredentialsListResponseAssetSpecAccessKey) GetKeyIdOk() (*int32, bool)`

GetKeyIdOk returns a tuple with the KeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKeyId

`func (o *CredentialsListResponseAssetSpecAccessKey) SetKeyId(v int32)`

SetKeyId sets KeyId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


