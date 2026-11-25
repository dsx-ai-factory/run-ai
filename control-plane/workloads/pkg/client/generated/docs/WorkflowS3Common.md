# WorkflowS3Common

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the profile, which tasks use to reference it. It must be unique among the profiles of its kind. (mandatory) | [optional] 
**Bucket** | Pointer to **NullableString** | The name of the bucket. (mandatory) | [optional] 
**Url** | Pointer to **NullableString** | The url of the S3 service provider. The default is the URL of the Amazon AWS S3 service. | [optional] 

## Methods

### NewWorkflowS3Common

`func NewWorkflowS3Common() *WorkflowS3Common`

NewWorkflowS3Common instantiates a new WorkflowS3Common object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowS3CommonWithDefaults

`func NewWorkflowS3CommonWithDefaults() *WorkflowS3Common`

NewWorkflowS3CommonWithDefaults instantiates a new WorkflowS3Common object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkflowS3Common) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkflowS3Common) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkflowS3Common) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkflowS3Common) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *WorkflowS3Common) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *WorkflowS3Common) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetBucket

`func (o *WorkflowS3Common) GetBucket() string`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *WorkflowS3Common) GetBucketOk() (*string, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *WorkflowS3Common) SetBucket(v string)`

SetBucket sets Bucket field to given value.

### HasBucket

`func (o *WorkflowS3Common) HasBucket() bool`

HasBucket returns a boolean if a field has been set.

### SetBucketNil

`func (o *WorkflowS3Common) SetBucketNil(b bool)`

 SetBucketNil sets the value for Bucket to be an explicit nil

### UnsetBucket
`func (o *WorkflowS3Common) UnsetBucket()`

UnsetBucket ensures that no value is present for Bucket, not even an explicit nil
### GetUrl

`func (o *WorkflowS3Common) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *WorkflowS3Common) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *WorkflowS3Common) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *WorkflowS3Common) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *WorkflowS3Common) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *WorkflowS3Common) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


