# ExtendedResourcesDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableExtendedResource**](ExtendedResource.md) |  | [optional] 
**Instances** | Pointer to [**[]ExtendedResource**](ExtendedResource.md) | Extended resources and their quantity. | [optional] 

## Methods

### NewExtendedResourcesDefaults

`func NewExtendedResourcesDefaults() *ExtendedResourcesDefaults`

NewExtendedResourcesDefaults instantiates a new ExtendedResourcesDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExtendedResourcesDefaultsWithDefaults

`func NewExtendedResourcesDefaultsWithDefaults() *ExtendedResourcesDefaults`

NewExtendedResourcesDefaultsWithDefaults instantiates a new ExtendedResourcesDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *ExtendedResourcesDefaults) GetAttributes() ExtendedResource`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *ExtendedResourcesDefaults) GetAttributesOk() (*ExtendedResource, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *ExtendedResourcesDefaults) SetAttributes(v ExtendedResource)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *ExtendedResourcesDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *ExtendedResourcesDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *ExtendedResourcesDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *ExtendedResourcesDefaults) GetInstances() []ExtendedResource`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *ExtendedResourcesDefaults) GetInstancesOk() (*[]ExtendedResource, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *ExtendedResourcesDefaults) SetInstances(v []ExtendedResource)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *ExtendedResourcesDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *ExtendedResourcesDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *ExtendedResourcesDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


