# S3Common

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bucket** | Pointer to **NullableString** | The name of the bucket. (mandatory) | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workload to which the S3 bucket will be mapped. (mandatory) | [optional] 
**Url** | Pointer to **NullableString** | The url of the S3 service provider. The default is the URL of the Amazon AWS S3 service. | [optional] 

## Methods

### NewS3Common

`func NewS3Common() *S3Common`

NewS3Common instantiates a new S3Common object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3CommonWithDefaults

`func NewS3CommonWithDefaults() *S3Common`

NewS3CommonWithDefaults instantiates a new S3Common object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBucket

`func (o *S3Common) GetBucket() string`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *S3Common) GetBucketOk() (*string, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *S3Common) SetBucket(v string)`

SetBucket sets Bucket field to given value.

### HasBucket

`func (o *S3Common) HasBucket() bool`

HasBucket returns a boolean if a field has been set.

### SetBucketNil

`func (o *S3Common) SetBucketNil(b bool)`

 SetBucketNil sets the value for Bucket to be an explicit nil

### UnsetBucket
`func (o *S3Common) UnsetBucket()`

UnsetBucket ensures that no value is present for Bucket, not even an explicit nil
### GetPath

`func (o *S3Common) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *S3Common) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *S3Common) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *S3Common) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *S3Common) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *S3Common) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetUrl

`func (o *S3Common) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *S3Common) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *S3Common) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *S3Common) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *S3Common) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *S3Common) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


