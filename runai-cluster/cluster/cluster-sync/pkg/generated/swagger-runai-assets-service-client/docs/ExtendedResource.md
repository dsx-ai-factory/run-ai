# ExtendedResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resource** | **string** | The name of the extended resource. | 
**Quantity** | **string** | The requested quantity for the given resource. | 
**Deleted** | Pointer to **bool** | Whether to exclude this extended resource from the workload. This is necessary in case the extended resource definition is inherited from a policy. | [optional] 

## Methods

### NewExtendedResource

`func NewExtendedResource(resource string, quantity string, ) *ExtendedResource`

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


### GetDeleted

`func (o *ExtendedResource) GetDeleted() bool`

GetDeleted returns the Deleted field if non-nil, zero value otherwise.

### GetDeletedOk

`func (o *ExtendedResource) GetDeletedOk() (*bool, bool)`

GetDeletedOk returns a tuple with the Deleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleted

`func (o *ExtendedResource) SetDeleted(v bool)`

SetDeleted sets Deleted field to given value.

### HasDeleted

`func (o *ExtendedResource) HasDeleted() bool`

HasDeleted returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


