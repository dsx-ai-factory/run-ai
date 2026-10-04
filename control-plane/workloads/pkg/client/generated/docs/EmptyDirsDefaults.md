# EmptyDirsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableEmptyDirInstance**](EmptyDirInstance.md) |  | [optional] 
**Instances** | Pointer to [**[]EmptyDirInstance**](EmptyDirInstance.md) | A list of emptyDir volumes to mount in the workload. | [optional] 

## Methods

### NewEmptyDirsDefaults

`func NewEmptyDirsDefaults() *EmptyDirsDefaults`

NewEmptyDirsDefaults instantiates a new EmptyDirsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEmptyDirsDefaultsWithDefaults

`func NewEmptyDirsDefaultsWithDefaults() *EmptyDirsDefaults`

NewEmptyDirsDefaultsWithDefaults instantiates a new EmptyDirsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *EmptyDirsDefaults) GetAttributes() EmptyDirInstance`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *EmptyDirsDefaults) GetAttributesOk() (*EmptyDirInstance, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *EmptyDirsDefaults) SetAttributes(v EmptyDirInstance)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *EmptyDirsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *EmptyDirsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *EmptyDirsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *EmptyDirsDefaults) GetInstances() []EmptyDirInstance`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *EmptyDirsDefaults) GetInstancesOk() (*[]EmptyDirInstance, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *EmptyDirsDefaults) SetInstances(v []EmptyDirInstance)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *EmptyDirsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *EmptyDirsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *EmptyDirsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


