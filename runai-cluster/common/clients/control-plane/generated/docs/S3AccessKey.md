# S3AccessKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessKeySecret** | Pointer to **NullableString** | Name of the secret containing credentials of the S3 bucket. Used for private S3 buckets. | [optional] 
**SecretKeyOfAccessKeyId** | Pointer to **NullableString** | The key to use for loading the access key id from the secret. The default is &#x60;AccessKeyId&#x60;. For more information, see [Credentials access keys](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html). | [optional] 
**SecretKeyOfSecretKey** | Pointer to **NullableString** | The key to use for loading the secret key from the secret. The default is &#x60;SecretKey&#x60;. For more information, see [Credentials access keys](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html). | [optional] 

## Methods

### NewS3AccessKey

`func NewS3AccessKey() *S3AccessKey`

NewS3AccessKey instantiates a new S3AccessKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3AccessKeyWithDefaults

`func NewS3AccessKeyWithDefaults() *S3AccessKey`

NewS3AccessKeyWithDefaults instantiates a new S3AccessKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessKeySecret

`func (o *S3AccessKey) GetAccessKeySecret() string`

GetAccessKeySecret returns the AccessKeySecret field if non-nil, zero value otherwise.

### GetAccessKeySecretOk

`func (o *S3AccessKey) GetAccessKeySecretOk() (*string, bool)`

GetAccessKeySecretOk returns a tuple with the AccessKeySecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeySecret

`func (o *S3AccessKey) SetAccessKeySecret(v string)`

SetAccessKeySecret sets AccessKeySecret field to given value.

### HasAccessKeySecret

`func (o *S3AccessKey) HasAccessKeySecret() bool`

HasAccessKeySecret returns a boolean if a field has been set.

### SetAccessKeySecretNil

`func (o *S3AccessKey) SetAccessKeySecretNil(b bool)`

 SetAccessKeySecretNil sets the value for AccessKeySecret to be an explicit nil

### UnsetAccessKeySecret
`func (o *S3AccessKey) UnsetAccessKeySecret()`

UnsetAccessKeySecret ensures that no value is present for AccessKeySecret, not even an explicit nil
### GetSecretKeyOfAccessKeyId

`func (o *S3AccessKey) GetSecretKeyOfAccessKeyId() string`

GetSecretKeyOfAccessKeyId returns the SecretKeyOfAccessKeyId field if non-nil, zero value otherwise.

### GetSecretKeyOfAccessKeyIdOk

`func (o *S3AccessKey) GetSecretKeyOfAccessKeyIdOk() (*string, bool)`

GetSecretKeyOfAccessKeyIdOk returns a tuple with the SecretKeyOfAccessKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfAccessKeyId

`func (o *S3AccessKey) SetSecretKeyOfAccessKeyId(v string)`

SetSecretKeyOfAccessKeyId sets SecretKeyOfAccessKeyId field to given value.

### HasSecretKeyOfAccessKeyId

`func (o *S3AccessKey) HasSecretKeyOfAccessKeyId() bool`

HasSecretKeyOfAccessKeyId returns a boolean if a field has been set.

### SetSecretKeyOfAccessKeyIdNil

`func (o *S3AccessKey) SetSecretKeyOfAccessKeyIdNil(b bool)`

 SetSecretKeyOfAccessKeyIdNil sets the value for SecretKeyOfAccessKeyId to be an explicit nil

### UnsetSecretKeyOfAccessKeyId
`func (o *S3AccessKey) UnsetSecretKeyOfAccessKeyId()`

UnsetSecretKeyOfAccessKeyId ensures that no value is present for SecretKeyOfAccessKeyId, not even an explicit nil
### GetSecretKeyOfSecretKey

`func (o *S3AccessKey) GetSecretKeyOfSecretKey() string`

GetSecretKeyOfSecretKey returns the SecretKeyOfSecretKey field if non-nil, zero value otherwise.

### GetSecretKeyOfSecretKeyOk

`func (o *S3AccessKey) GetSecretKeyOfSecretKeyOk() (*string, bool)`

GetSecretKeyOfSecretKeyOk returns a tuple with the SecretKeyOfSecretKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfSecretKey

`func (o *S3AccessKey) SetSecretKeyOfSecretKey(v string)`

SetSecretKeyOfSecretKey sets SecretKeyOfSecretKey field to given value.

### HasSecretKeyOfSecretKey

`func (o *S3AccessKey) HasSecretKeyOfSecretKey() bool`

HasSecretKeyOfSecretKey returns a boolean if a field has been set.

### SetSecretKeyOfSecretKeyNil

`func (o *S3AccessKey) SetSecretKeyOfSecretKeyNil(b bool)`

 SetSecretKeyOfSecretKeyNil sets the value for SecretKeyOfSecretKey to be an explicit nil

### UnsetSecretKeyOfSecretKey
`func (o *S3AccessKey) UnsetSecretKeyOfSecretKey()`

UnsetSecretKeyOfSecretKey ensures that no value is present for SecretKeyOfSecretKey, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


