# AcceleratorDevice

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | **string** | Exact Kubernetes node label key for this device model. | 
**Value** | **string** | Exact Kubernetes node label value for this device model. | 

## Methods

### NewAcceleratorDevice

`func NewAcceleratorDevice(key string, value string, ) *AcceleratorDevice`

NewAcceleratorDevice instantiates a new AcceleratorDevice object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAcceleratorDeviceWithDefaults

`func NewAcceleratorDeviceWithDefaults() *AcceleratorDevice`

NewAcceleratorDeviceWithDefaults instantiates a new AcceleratorDevice object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *AcceleratorDevice) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *AcceleratorDevice) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *AcceleratorDevice) SetKey(v string)`

SetKey sets Key field to given value.


### GetValue

`func (o *AcceleratorDevice) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *AcceleratorDevice) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *AcceleratorDevice) SetValue(v string)`

SetValue sets Value field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


