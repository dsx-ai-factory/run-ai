# S3AssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bucket** | Pointer to **NullableString** | The name of the bucket. (mandatory) | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workload to which the S3 bucket will be mapped. (mandatory) | [optional] 
**Url** | Pointer to **NullableString** | The url of the S3 service provider. The default is the URL of the Amazon AWS S3 service. | [optional] 
**AccessKeyAssetId** | Pointer to **NullableString** | ID of credentials asset of type access-key, for private S3 buckets. | [optional] 

## Methods

### NewS3AssetSpec

`func NewS3AssetSpec() *S3AssetSpec`

NewS3AssetSpec instantiates a new S3AssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3AssetSpecWithDefaults

`func NewS3AssetSpecWithDefaults() *S3AssetSpec`

NewS3AssetSpecWithDefaults instantiates a new S3AssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBucket

`func (o *S3AssetSpec) GetBucket() string`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *S3AssetSpec) GetBucketOk() (*string, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *S3AssetSpec) SetBucket(v string)`

SetBucket sets Bucket field to given value.

### HasBucket

`func (o *S3AssetSpec) HasBucket() bool`

HasBucket returns a boolean if a field has been set.

### SetBucketNil

`func (o *S3AssetSpec) SetBucketNil(b bool)`

 SetBucketNil sets the value for Bucket to be an explicit nil

### UnsetBucket
`func (o *S3AssetSpec) UnsetBucket()`

UnsetBucket ensures that no value is present for Bucket, not even an explicit nil
### GetPath

`func (o *S3AssetSpec) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *S3AssetSpec) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *S3AssetSpec) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *S3AssetSpec) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *S3AssetSpec) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *S3AssetSpec) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetUrl

`func (o *S3AssetSpec) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *S3AssetSpec) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *S3AssetSpec) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *S3AssetSpec) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *S3AssetSpec) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *S3AssetSpec) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil
### GetAccessKeyAssetId

`func (o *S3AssetSpec) GetAccessKeyAssetId() string`

GetAccessKeyAssetId returns the AccessKeyAssetId field if non-nil, zero value otherwise.

### GetAccessKeyAssetIdOk

`func (o *S3AssetSpec) GetAccessKeyAssetIdOk() (*string, bool)`

GetAccessKeyAssetIdOk returns a tuple with the AccessKeyAssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeyAssetId

`func (o *S3AssetSpec) SetAccessKeyAssetId(v string)`

SetAccessKeyAssetId sets AccessKeyAssetId field to given value.

### HasAccessKeyAssetId

`func (o *S3AssetSpec) HasAccessKeyAssetId() bool`

HasAccessKeyAssetId returns a boolean if a field has been set.

### SetAccessKeyAssetIdNil

`func (o *S3AssetSpec) SetAccessKeyAssetIdNil(b bool)`

 SetAccessKeyAssetIdNil sets the value for AccessKeyAssetId to be an explicit nil

### UnsetAccessKeyAssetId
`func (o *S3AssetSpec) UnsetAccessKeyAssetId()`

UnsetAccessKeyAssetId ensures that no value is present for AccessKeyAssetId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


