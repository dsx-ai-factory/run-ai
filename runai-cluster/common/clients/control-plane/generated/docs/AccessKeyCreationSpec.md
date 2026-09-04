# AccessKeyCreationSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | optional name of existing secret in the cluster from which the access key id and secret should be taken. If omitted, you will have to provide those credentials in the accessKeyId and secretAccessKey fields. The provided credentials are encrypted into the control plane database and cloned to a kubernetes secret in the cluster. | [optional] 
**AccessKeyId** | Pointer to **NullableString** | The access key id of the S3-compatible bucket, required only when not using existing secret. | [optional] 
**SecretAccessKey** | Pointer to **NullableString** | The secret access key of the S3-compatible bucket, required only when not using existing secret. | [optional] 

## Methods

### NewAccessKeyCreationSpec

`func NewAccessKeyCreationSpec() *AccessKeyCreationSpec`

NewAccessKeyCreationSpec instantiates a new AccessKeyCreationSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessKeyCreationSpecWithDefaults

`func NewAccessKeyCreationSpecWithDefaults() *AccessKeyCreationSpec`

NewAccessKeyCreationSpecWithDefaults instantiates a new AccessKeyCreationSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *AccessKeyCreationSpec) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *AccessKeyCreationSpec) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *AccessKeyCreationSpec) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *AccessKeyCreationSpec) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *AccessKeyCreationSpec) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *AccessKeyCreationSpec) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetAccessKeyId

`func (o *AccessKeyCreationSpec) GetAccessKeyId() string`

GetAccessKeyId returns the AccessKeyId field if non-nil, zero value otherwise.

### GetAccessKeyIdOk

`func (o *AccessKeyCreationSpec) GetAccessKeyIdOk() (*string, bool)`

GetAccessKeyIdOk returns a tuple with the AccessKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeyId

`func (o *AccessKeyCreationSpec) SetAccessKeyId(v string)`

SetAccessKeyId sets AccessKeyId field to given value.

### HasAccessKeyId

`func (o *AccessKeyCreationSpec) HasAccessKeyId() bool`

HasAccessKeyId returns a boolean if a field has been set.

### SetAccessKeyIdNil

`func (o *AccessKeyCreationSpec) SetAccessKeyIdNil(b bool)`

 SetAccessKeyIdNil sets the value for AccessKeyId to be an explicit nil

### UnsetAccessKeyId
`func (o *AccessKeyCreationSpec) UnsetAccessKeyId()`

UnsetAccessKeyId ensures that no value is present for AccessKeyId, not even an explicit nil
### GetSecretAccessKey

`func (o *AccessKeyCreationSpec) GetSecretAccessKey() string`

GetSecretAccessKey returns the SecretAccessKey field if non-nil, zero value otherwise.

### GetSecretAccessKeyOk

`func (o *AccessKeyCreationSpec) GetSecretAccessKeyOk() (*string, bool)`

GetSecretAccessKeyOk returns a tuple with the SecretAccessKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretAccessKey

`func (o *AccessKeyCreationSpec) SetSecretAccessKey(v string)`

SetSecretAccessKey sets SecretAccessKey field to given value.

### HasSecretAccessKey

`func (o *AccessKeyCreationSpec) HasSecretAccessKey() bool`

HasSecretAccessKey returns a boolean if a field has been set.

### SetSecretAccessKeyNil

`func (o *AccessKeyCreationSpec) SetSecretAccessKeyNil(b bool)`

 SetSecretAccessKeyNil sets the value for SecretAccessKey to be an explicit nil

### UnsetSecretAccessKey
`func (o *AccessKeyCreationSpec) UnsetSecretAccessKey()`

UnsetSecretAccessKey ensures that no value is present for SecretAccessKey, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


