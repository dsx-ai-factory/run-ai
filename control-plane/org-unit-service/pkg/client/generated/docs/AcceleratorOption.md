# AcceleratorOption

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceName** | **string** | Allocatable Kubernetes extended resource name. | 
**Devices** | [**[]AcceleratorDevice**](AcceleratorDevice.md) | Model selectors observed alongside this resource on the same node. | 

## Methods

### NewAcceleratorOption

`func NewAcceleratorOption(resourceName string, devices []AcceleratorDevice, ) *AcceleratorOption`

NewAcceleratorOption instantiates a new AcceleratorOption object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAcceleratorOptionWithDefaults

`func NewAcceleratorOptionWithDefaults() *AcceleratorOption`

NewAcceleratorOptionWithDefaults instantiates a new AcceleratorOption object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceName

`func (o *AcceleratorOption) GetResourceName() string`

GetResourceName returns the ResourceName field if non-nil, zero value otherwise.

### GetResourceNameOk

`func (o *AcceleratorOption) GetResourceNameOk() (*string, bool)`

GetResourceNameOk returns a tuple with the ResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceName

`func (o *AcceleratorOption) SetResourceName(v string)`

SetResourceName sets ResourceName field to given value.


### GetDevices

`func (o *AcceleratorOption) GetDevices() []AcceleratorDevice`

GetDevices returns the Devices field if non-nil, zero value otherwise.

### GetDevicesOk

`func (o *AcceleratorOption) GetDevicesOk() (*[]AcceleratorDevice, bool)`

GetDevicesOk returns a tuple with the Devices field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevices

`func (o *AcceleratorOption) SetDevices(v []AcceleratorDevice)`

SetDevices sets Devices field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


