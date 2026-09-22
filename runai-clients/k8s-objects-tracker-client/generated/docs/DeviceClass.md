# DeviceClass

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | Cluster ID which has sent the resource info | [optional] 
**DeviceClassName** | **string** | Device class name | 

## Methods

### NewDeviceClass

`func NewDeviceClass(deviceClassName string, ) *DeviceClass`

NewDeviceClass instantiates a new DeviceClass object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeviceClassWithDefaults

`func NewDeviceClassWithDefaults() *DeviceClass`

NewDeviceClassWithDefaults instantiates a new DeviceClass object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *DeviceClass) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DeviceClass) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DeviceClass) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *DeviceClass) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetDeviceClassName

`func (o *DeviceClass) GetDeviceClassName() string`

GetDeviceClassName returns the DeviceClassName field if non-nil, zero value otherwise.

### GetDeviceClassNameOk

`func (o *DeviceClass) GetDeviceClassNameOk() (*string, bool)`

GetDeviceClassNameOk returns a tuple with the DeviceClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeviceClassName

`func (o *DeviceClass) SetDeviceClassName(v string)`

SetDeviceClassName sets DeviceClassName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


