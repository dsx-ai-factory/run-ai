# WorkloadsExtendedResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resource** | Pointer to **NullableString** | The name of the extended resource (mandatory) | [optional] 
**Quantity** | Pointer to **NullableString** | The requested quantity for the given resource. | [optional] 
**Exclude** | Pointer to **NullableBool** | Whether to exclude this extended resource from the workload. This is necessary in case the extended resource is inherited from the policy defaults and it is desired not to include it in this workload. | [optional] 

## Methods

### NewWorkloadsExtendedResource

`func NewWorkloadsExtendedResource() *WorkloadsExtendedResource`

NewWorkloadsExtendedResource instantiates a new WorkloadsExtendedResource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadsExtendedResourceWithDefaults

`func NewWorkloadsExtendedResourceWithDefaults() *WorkloadsExtendedResource`

NewWorkloadsExtendedResourceWithDefaults instantiates a new WorkloadsExtendedResource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResource

`func (o *WorkloadsExtendedResource) GetResource() string`

GetResource returns the Resource field if non-nil, zero value otherwise.

### GetResourceOk

`func (o *WorkloadsExtendedResource) GetResourceOk() (*string, bool)`

GetResourceOk returns a tuple with the Resource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResource

`func (o *WorkloadsExtendedResource) SetResource(v string)`

SetResource sets Resource field to given value.

### HasResource

`func (o *WorkloadsExtendedResource) HasResource() bool`

HasResource returns a boolean if a field has been set.

### SetResourceNil

`func (o *WorkloadsExtendedResource) SetResourceNil(b bool)`

 SetResourceNil sets the value for Resource to be an explicit nil

### UnsetResource
`func (o *WorkloadsExtendedResource) UnsetResource()`

UnsetResource ensures that no value is present for Resource, not even an explicit nil
### GetQuantity

`func (o *WorkloadsExtendedResource) GetQuantity() string`

GetQuantity returns the Quantity field if non-nil, zero value otherwise.

### GetQuantityOk

`func (o *WorkloadsExtendedResource) GetQuantityOk() (*string, bool)`

GetQuantityOk returns a tuple with the Quantity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuantity

`func (o *WorkloadsExtendedResource) SetQuantity(v string)`

SetQuantity sets Quantity field to given value.

### HasQuantity

`func (o *WorkloadsExtendedResource) HasQuantity() bool`

HasQuantity returns a boolean if a field has been set.

### SetQuantityNil

`func (o *WorkloadsExtendedResource) SetQuantityNil(b bool)`

 SetQuantityNil sets the value for Quantity to be an explicit nil

### UnsetQuantity
`func (o *WorkloadsExtendedResource) UnsetQuantity()`

UnsetQuantity ensures that no value is present for Quantity, not even an explicit nil
### GetExclude

`func (o *WorkloadsExtendedResource) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *WorkloadsExtendedResource) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *WorkloadsExtendedResource) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *WorkloadsExtendedResource) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *WorkloadsExtendedResource) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *WorkloadsExtendedResource) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


