# AccessKeyUpdateSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | optional name of existing secret in the cluster from which the access key id and secret should be taken. If omitted, you will have to provide those credentials in the accessKeyId and secretAccessKey fields. The provided credentials are encrypted into the control plane database and cloned to a kubernetes secret in the cluster. | [optional] 
**AccessKeyId** | Pointer to **NullableString** | The access key id of the S3-compatible bucket, required only when not using existing secret. | [optional] 
**SecretAccessKey** | Pointer to **NullableString** | The secret access key of the S3-compatible bucket, required only when not using existing secret. | [optional] 

## Methods

### NewAccessKeyUpdateSpec

`func NewAccessKeyUpdateSpec() *AccessKeyUpdateSpec`

NewAccessKeyUpdateSpec instantiates a new AccessKeyUpdateSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessKeyUpdateSpecWithDefaults

`func NewAccessKeyUpdateSpecWithDefaults() *AccessKeyUpdateSpec`

NewAccessKeyUpdateSpecWithDefaults instantiates a new AccessKeyUpdateSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *AccessKeyUpdateSpec) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *AccessKeyUpdateSpec) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *AccessKeyUpdateSpec) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *AccessKeyUpdateSpec) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *AccessKeyUpdateSpec) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *AccessKeyUpdateSpec) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetAccessKeyId

`func (o *AccessKeyUpdateSpec) GetAccessKeyId() string`

GetAccessKeyId returns the AccessKeyId field if non-nil, zero value otherwise.

### GetAccessKeyIdOk

`func (o *AccessKeyUpdateSpec) GetAccessKeyIdOk() (*string, bool)`

GetAccessKeyIdOk returns a tuple with the AccessKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeyId

`func (o *AccessKeyUpdateSpec) SetAccessKeyId(v string)`

SetAccessKeyId sets AccessKeyId field to given value.

### HasAccessKeyId

`func (o *AccessKeyUpdateSpec) HasAccessKeyId() bool`

HasAccessKeyId returns a boolean if a field has been set.

### SetAccessKeyIdNil

`func (o *AccessKeyUpdateSpec) SetAccessKeyIdNil(b bool)`

 SetAccessKeyIdNil sets the value for AccessKeyId to be an explicit nil

### UnsetAccessKeyId
`func (o *AccessKeyUpdateSpec) UnsetAccessKeyId()`

UnsetAccessKeyId ensures that no value is present for AccessKeyId, not even an explicit nil
### GetSecretAccessKey

`func (o *AccessKeyUpdateSpec) GetSecretAccessKey() string`

GetSecretAccessKey returns the SecretAccessKey field if non-nil, zero value otherwise.

### GetSecretAccessKeyOk

`func (o *AccessKeyUpdateSpec) GetSecretAccessKeyOk() (*string, bool)`

GetSecretAccessKeyOk returns a tuple with the SecretAccessKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretAccessKey

`func (o *AccessKeyUpdateSpec) SetSecretAccessKey(v string)`

SetSecretAccessKey sets SecretAccessKey field to given value.

### HasSecretAccessKey

`func (o *AccessKeyUpdateSpec) HasSecretAccessKey() bool`

HasSecretAccessKey returns a boolean if a field has been set.

### SetSecretAccessKeyNil

`func (o *AccessKeyUpdateSpec) SetSecretAccessKeyNil(b bool)`

 SetSecretAccessKeyNil sets the value for SecretAccessKey to be an explicit nil

### UnsetSecretAccessKey
`func (o *AccessKeyUpdateSpec) UnsetSecretAccessKey()`

UnsetSecretAccessKey ensures that no value is present for SecretAccessKey, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


