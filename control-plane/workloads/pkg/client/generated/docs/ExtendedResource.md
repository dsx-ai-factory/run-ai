# ExtendedResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resource** | Pointer to **NullableString** | The name of the extended resource (mandatory) | [optional] 
**Quantity** | Pointer to **NullableString** | The requested quantity for the resource. | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the extended resource is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] 

## Methods

### NewExtendedResource

`func NewExtendedResource() *ExtendedResource`

NewExtendedResource instantiates a new ExtendedResource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExtendedResourceWithDefaults

`func NewExtendedResourceWithDefaults() *ExtendedResource`

NewExtendedResourceWithDefaults instantiates a new ExtendedResource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResource

`func (o *ExtendedResource) GetResource() string`

GetResource returns the Resource field if non-nil, zero value otherwise.

### GetResourceOk

`func (o *ExtendedResource) GetResourceOk() (*string, bool)`

GetResourceOk returns a tuple with the Resource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResource

`func (o *ExtendedResource) SetResource(v string)`

SetResource sets Resource field to given value.

### HasResource

`func (o *ExtendedResource) HasResource() bool`

HasResource returns a boolean if a field has been set.

### SetResourceNil

`func (o *ExtendedResource) SetResourceNil(b bool)`

 SetResourceNil sets the value for Resource to be an explicit nil

### UnsetResource
`func (o *ExtendedResource) UnsetResource()`

UnsetResource ensures that no value is present for Resource, not even an explicit nil
### GetQuantity

`func (o *ExtendedResource) GetQuantity() string`

GetQuantity returns the Quantity field if non-nil, zero value otherwise.

### GetQuantityOk

`func (o *ExtendedResource) GetQuantityOk() (*string, bool)`

GetQuantityOk returns a tuple with the Quantity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuantity

`func (o *ExtendedResource) SetQuantity(v string)`

SetQuantity sets Quantity field to given value.

### HasQuantity

`func (o *ExtendedResource) HasQuantity() bool`

HasQuantity returns a boolean if a field has been set.

### SetQuantityNil

`func (o *ExtendedResource) SetQuantityNil(b bool)`

 SetQuantityNil sets the value for Quantity to be an explicit nil

### UnsetQuantity
`func (o *ExtendedResource) UnsetQuantity()`

UnsetQuantity ensures that no value is present for Quantity, not even an explicit nil
### GetExclude

`func (o *ExtendedResource) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *ExtendedResource) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *ExtendedResource) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *ExtendedResource) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *ExtendedResource) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *ExtendedResource) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


