# UpdateWorkloadsSettingByKeyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** | Whether to enable the model catalog (Hugging Face models). | [optional] 

## Methods

### NewUpdateWorkloadsSettingByKeyRequest

`func NewUpdateWorkloadsSettingByKeyRequest() *UpdateWorkloadsSettingByKeyRequest`

NewUpdateWorkloadsSettingByKeyRequest instantiates a new UpdateWorkloadsSettingByKeyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateWorkloadsSettingByKeyRequestWithDefaults

`func NewUpdateWorkloadsSettingByKeyRequestWithDefaults() *UpdateWorkloadsSettingByKeyRequest`

NewUpdateWorkloadsSettingByKeyRequestWithDefaults instantiates a new UpdateWorkloadsSettingByKeyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *UpdateWorkloadsSettingByKeyRequest) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *UpdateWorkloadsSettingByKeyRequest) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *UpdateWorkloadsSettingByKeyRequest) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *UpdateWorkloadsSettingByKeyRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *UpdateWorkloadsSettingByKeyRequest) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *UpdateWorkloadsSettingByKeyRequest) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


