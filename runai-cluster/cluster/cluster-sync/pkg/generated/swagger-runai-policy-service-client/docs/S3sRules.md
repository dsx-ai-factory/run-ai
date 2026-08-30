# S3sRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableS3Rules**](S3Rules.md) |  | [optional] 
**Instances** | Pointer to [**NullableItemRules**](ItemRules.md) |  | [optional] 

## Methods

### NewS3sRules

`func NewS3sRules() *S3sRules`

NewS3sRules instantiates a new S3sRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3sRulesWithDefaults

`func NewS3sRulesWithDefaults() *S3sRules`

NewS3sRulesWithDefaults instantiates a new S3sRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *S3sRules) GetAttributes() S3Rules`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *S3sRules) GetAttributesOk() (*S3Rules, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *S3sRules) SetAttributes(v S3Rules)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *S3sRules) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *S3sRules) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *S3sRules) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *S3sRules) GetInstances() ItemRules`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *S3sRules) GetInstancesOk() (*ItemRules, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *S3sRules) SetInstances(v ItemRules)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *S3sRules) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *S3sRules) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *S3sRules) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


