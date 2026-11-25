# AcceleratorSelection

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceName** | Pointer to **NullableString** | Kubernetes extended resource name. Defaults to nvidia.com/gpu. | [optional] 
**DeviceLabelKey** | Pointer to **NullableString** | Node label key for gpuType. Defaults to nvidia.com/gpu.product for NVIDIA. | [optional] 

## Methods

### NewAcceleratorSelection

`func NewAcceleratorSelection() *AcceleratorSelection`

NewAcceleratorSelection instantiates a new AcceleratorSelection object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAcceleratorSelectionWithDefaults

`func NewAcceleratorSelectionWithDefaults() *AcceleratorSelection`

NewAcceleratorSelectionWithDefaults instantiates a new AcceleratorSelection object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceName

`func (o *AcceleratorSelection) GetResourceName() string`

GetResourceName returns the ResourceName field if non-nil, zero value otherwise.

### GetResourceNameOk

`func (o *AcceleratorSelection) GetResourceNameOk() (*string, bool)`

GetResourceNameOk returns a tuple with the ResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceName

`func (o *AcceleratorSelection) SetResourceName(v string)`

SetResourceName sets ResourceName field to given value.

### HasResourceName

`func (o *AcceleratorSelection) HasResourceName() bool`

HasResourceName returns a boolean if a field has been set.

### SetResourceNameNil

`func (o *AcceleratorSelection) SetResourceNameNil(b bool)`

 SetResourceNameNil sets the value for ResourceName to be an explicit nil

### UnsetResourceName
`func (o *AcceleratorSelection) UnsetResourceName()`

UnsetResourceName ensures that no value is present for ResourceName, not even an explicit nil
### GetDeviceLabelKey

`func (o *AcceleratorSelection) GetDeviceLabelKey() string`

GetDeviceLabelKey returns the DeviceLabelKey field if non-nil, zero value otherwise.

### GetDeviceLabelKeyOk

`func (o *AcceleratorSelection) GetDeviceLabelKeyOk() (*string, bool)`

GetDeviceLabelKeyOk returns a tuple with the DeviceLabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceLabelKey

`func (o *AcceleratorSelection) SetDeviceLabelKey(v string)`

SetDeviceLabelKey sets DeviceLabelKey field to given value.

### HasDeviceLabelKey

`func (o *AcceleratorSelection) HasDeviceLabelKey() bool`

HasDeviceLabelKey returns a boolean if a field has been set.

### SetDeviceLabelKeyNil

`func (o *AcceleratorSelection) SetDeviceLabelKeyNil(b bool)`

 SetDeviceLabelKeyNil sets the value for DeviceLabelKey to be an explicit nil

### UnsetDeviceLabelKey
`func (o *AcceleratorSelection) UnsetDeviceLabelKey()`

UnsetDeviceLabelKey ensures that no value is present for DeviceLabelKey, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


