# S3Rules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bucket** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Path** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Url** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewS3Rules

`func NewS3Rules() *S3Rules`

NewS3Rules instantiates a new S3Rules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3RulesWithDefaults

`func NewS3RulesWithDefaults() *S3Rules`

NewS3RulesWithDefaults instantiates a new S3Rules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBucket

`func (o *S3Rules) GetBucket() StringRules`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *S3Rules) GetBucketOk() (*StringRules, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *S3Rules) SetBucket(v StringRules)`

SetBucket sets Bucket field to given value.

### HasBucket

`func (o *S3Rules) HasBucket() bool`

HasBucket returns a boolean if a field has been set.

### SetBucketNil

`func (o *S3Rules) SetBucketNil(b bool)`

 SetBucketNil sets the value for Bucket to be an explicit nil

### UnsetBucket
`func (o *S3Rules) UnsetBucket()`

UnsetBucket ensures that no value is present for Bucket, not even an explicit nil
### GetPath

`func (o *S3Rules) GetPath() StringRules`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *S3Rules) GetPathOk() (*StringRules, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *S3Rules) SetPath(v StringRules)`

SetPath sets Path field to given value.

### HasPath

`func (o *S3Rules) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *S3Rules) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *S3Rules) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetUrl

`func (o *S3Rules) GetUrl() StringRules`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *S3Rules) GetUrlOk() (*StringRules, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *S3Rules) SetUrl(v StringRules)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *S3Rules) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *S3Rules) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *S3Rules) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


