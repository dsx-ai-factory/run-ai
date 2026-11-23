# AcceleratorSelection

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceName** | **string** |  | 
**DeviceLabelKey** | **string** |  | 
**GpuType** | **string** |  | 

## Methods

### NewAcceleratorSelection

`func NewAcceleratorSelection(resourceName string, deviceLabelKey string, gpuType string, ) *AcceleratorSelection`

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


### GetGpuType

`func (o *AcceleratorSelection) GetGpuType() string`

GetGpuType returns the GpuType field if non-nil, zero value otherwise.

### GetGpuTypeOk

`func (o *AcceleratorSelection) GetGpuTypeOk() (*string, bool)`

GetGpuTypeOk returns a tuple with the GpuType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuType

`func (o *AcceleratorSelection) SetGpuType(v string)`

SetGpuType sets GpuType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


