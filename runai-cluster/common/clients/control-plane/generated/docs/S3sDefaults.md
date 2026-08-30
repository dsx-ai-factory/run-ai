# S3sDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableS3Instance**](S3Instance.md) |  | [optional] 
**Instances** | Pointer to [**[]S3Instance**](S3Instance.md) | Set of s3 buckets to use in the workload. | [optional] 

## Methods

### NewS3sDefaults

`func NewS3sDefaults() *S3sDefaults`

NewS3sDefaults instantiates a new S3sDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3sDefaultsWithDefaults

`func NewS3sDefaultsWithDefaults() *S3sDefaults`

NewS3sDefaultsWithDefaults instantiates a new S3sDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *S3sDefaults) GetAttributes() S3Instance`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *S3sDefaults) GetAttributesOk() (*S3Instance, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *S3sDefaults) SetAttributes(v S3Instance)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *S3sDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *S3sDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *S3sDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *S3sDefaults) GetInstances() []S3Instance`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *S3sDefaults) GetInstancesOk() (*[]S3Instance, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *S3sDefaults) SetInstances(v []S3Instance)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *S3sDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *S3sDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *S3sDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


