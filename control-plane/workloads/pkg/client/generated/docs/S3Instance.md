# S3Instance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Bucket** | Pointer to **NullableString** | The name of the bucket. (mandatory) | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workload to which the S3 bucket will be mapped. (mandatory) | [optional] 
**Url** | Pointer to **NullableString** | The url of the S3 service provider. The default is the URL of the Amazon AWS S3 service. | [optional] 
**AccessKeySecret** | Pointer to **NullableString** | Name of the secret containing credentials of the S3 bucket. Used for private S3 buckets. | [optional] 
**SecretKeyOfAccessKeyId** | Pointer to **NullableString** | The key to use for loading the access key id from the secret. The default is &#x60;AccessKeyId&#x60;. For more information, see [Credentials access keys](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html). | [optional] 
**SecretKeyOfSecretKey** | Pointer to **NullableString** | The key to use for loading the secret key from the secret. The default is &#x60;SecretKey&#x60;. For more information, see [Credentials access keys](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html). | [optional] 
**AwsSessionToken** | Pointer to **NullableString** | An AWS session token. It should be provided (in addition to the access key ID and secret access key) to authenticate to a bucket that is configured to use AWS Security Token Service (STS). | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewS3Instance

`func NewS3Instance() *S3Instance`

NewS3Instance instantiates a new S3Instance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3InstanceWithDefaults

`func NewS3InstanceWithDefaults() *S3Instance`

NewS3InstanceWithDefaults instantiates a new S3Instance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *S3Instance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *S3Instance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *S3Instance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *S3Instance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *S3Instance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *S3Instance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetBucket

`func (o *S3Instance) GetBucket() string`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *S3Instance) GetBucketOk() (*string, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *S3Instance) SetBucket(v string)`

SetBucket sets Bucket field to given value.

### HasBucket

`func (o *S3Instance) HasBucket() bool`

HasBucket returns a boolean if a field has been set.

### SetBucketNil

`func (o *S3Instance) SetBucketNil(b bool)`

 SetBucketNil sets the value for Bucket to be an explicit nil

### UnsetBucket
`func (o *S3Instance) UnsetBucket()`

UnsetBucket ensures that no value is present for Bucket, not even an explicit nil
### GetPath

`func (o *S3Instance) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *S3Instance) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *S3Instance) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *S3Instance) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *S3Instance) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *S3Instance) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetUrl

`func (o *S3Instance) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *S3Instance) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *S3Instance) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *S3Instance) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *S3Instance) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *S3Instance) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil
### GetAccessKeySecret

`func (o *S3Instance) GetAccessKeySecret() string`

GetAccessKeySecret returns the AccessKeySecret field if non-nil, zero value otherwise.

### GetAccessKeySecretOk

`func (o *S3Instance) GetAccessKeySecretOk() (*string, bool)`

GetAccessKeySecretOk returns a tuple with the AccessKeySecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeySecret

`func (o *S3Instance) SetAccessKeySecret(v string)`

SetAccessKeySecret sets AccessKeySecret field to given value.

### HasAccessKeySecret

`func (o *S3Instance) HasAccessKeySecret() bool`

HasAccessKeySecret returns a boolean if a field has been set.

### SetAccessKeySecretNil

`func (o *S3Instance) SetAccessKeySecretNil(b bool)`

 SetAccessKeySecretNil sets the value for AccessKeySecret to be an explicit nil

### UnsetAccessKeySecret
`func (o *S3Instance) UnsetAccessKeySecret()`

UnsetAccessKeySecret ensures that no value is present for AccessKeySecret, not even an explicit nil
### GetSecretKeyOfAccessKeyId

`func (o *S3Instance) GetSecretKeyOfAccessKeyId() string`

GetSecretKeyOfAccessKeyId returns the SecretKeyOfAccessKeyId field if non-nil, zero value otherwise.

### GetSecretKeyOfAccessKeyIdOk

`func (o *S3Instance) GetSecretKeyOfAccessKeyIdOk() (*string, bool)`

GetSecretKeyOfAccessKeyIdOk returns a tuple with the SecretKeyOfAccessKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfAccessKeyId

`func (o *S3Instance) SetSecretKeyOfAccessKeyId(v string)`

SetSecretKeyOfAccessKeyId sets SecretKeyOfAccessKeyId field to given value.

### HasSecretKeyOfAccessKeyId

`func (o *S3Instance) HasSecretKeyOfAccessKeyId() bool`

HasSecretKeyOfAccessKeyId returns a boolean if a field has been set.

### SetSecretKeyOfAccessKeyIdNil

`func (o *S3Instance) SetSecretKeyOfAccessKeyIdNil(b bool)`

 SetSecretKeyOfAccessKeyIdNil sets the value for SecretKeyOfAccessKeyId to be an explicit nil

### UnsetSecretKeyOfAccessKeyId
`func (o *S3Instance) UnsetSecretKeyOfAccessKeyId()`

UnsetSecretKeyOfAccessKeyId ensures that no value is present for SecretKeyOfAccessKeyId, not even an explicit nil
### GetSecretKeyOfSecretKey

`func (o *S3Instance) GetSecretKeyOfSecretKey() string`

GetSecretKeyOfSecretKey returns the SecretKeyOfSecretKey field if non-nil, zero value otherwise.

### GetSecretKeyOfSecretKeyOk

`func (o *S3Instance) GetSecretKeyOfSecretKeyOk() (*string, bool)`

GetSecretKeyOfSecretKeyOk returns a tuple with the SecretKeyOfSecretKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfSecretKey

`func (o *S3Instance) SetSecretKeyOfSecretKey(v string)`

SetSecretKeyOfSecretKey sets SecretKeyOfSecretKey field to given value.

### HasSecretKeyOfSecretKey

`func (o *S3Instance) HasSecretKeyOfSecretKey() bool`

HasSecretKeyOfSecretKey returns a boolean if a field has been set.

### SetSecretKeyOfSecretKeyNil

`func (o *S3Instance) SetSecretKeyOfSecretKeyNil(b bool)`

 SetSecretKeyOfSecretKeyNil sets the value for SecretKeyOfSecretKey to be an explicit nil

### UnsetSecretKeyOfSecretKey
`func (o *S3Instance) UnsetSecretKeyOfSecretKey()`

UnsetSecretKeyOfSecretKey ensures that no value is present for SecretKeyOfSecretKey, not even an explicit nil
### GetAwsSessionToken

`func (o *S3Instance) GetAwsSessionToken() string`

GetAwsSessionToken returns the AwsSessionToken field if non-nil, zero value otherwise.

### GetAwsSessionTokenOk

`func (o *S3Instance) GetAwsSessionTokenOk() (*string, bool)`

GetAwsSessionTokenOk returns a tuple with the AwsSessionToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAwsSessionToken

`func (o *S3Instance) SetAwsSessionToken(v string)`

SetAwsSessionToken sets AwsSessionToken field to given value.

### HasAwsSessionToken

`func (o *S3Instance) HasAwsSessionToken() bool`

HasAwsSessionToken returns a boolean if a field has been set.

### SetAwsSessionTokenNil

`func (o *S3Instance) SetAwsSessionTokenNil(b bool)`

 SetAwsSessionTokenNil sets the value for AwsSessionToken to be an explicit nil

### UnsetAwsSessionToken
`func (o *S3Instance) UnsetAwsSessionToken()`

UnsetAwsSessionToken ensures that no value is present for AwsSessionToken, not even an explicit nil
### GetExclude

`func (o *S3Instance) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *S3Instance) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *S3Instance) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *S3Instance) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *S3Instance) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *S3Instance) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


