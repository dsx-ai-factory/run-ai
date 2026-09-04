# NodePoolSetting

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** |  | [optional] 
**DisplayName** | Pointer to **string** |  | [optional] [readonly] 
**Description** | Pointer to **string** |  | [optional] [readonly] 
**Stage** | Pointer to **string** |  | [optional] [readonly] 

## Methods

### NewNodePoolSetting

`func NewNodePoolSetting() *NodePoolSetting`

NewNodePoolSetting instantiates a new NodePoolSetting object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolSettingWithDefaults

`func NewNodePoolSettingWithDefaults() *NodePoolSetting`

NewNodePoolSettingWithDefaults instantiates a new NodePoolSetting object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *NodePoolSetting) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *NodePoolSetting) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *NodePoolSetting) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *NodePoolSetting) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *NodePoolSetting) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *NodePoolSetting) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetDisplayName

`func (o *NodePoolSetting) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *NodePoolSetting) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *NodePoolSetting) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *NodePoolSetting) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *NodePoolSetting) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *NodePoolSetting) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *NodePoolSetting) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *NodePoolSetting) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetStage

`func (o *NodePoolSetting) GetStage() string`

GetStage returns the Stage field if non-nil, zero value otherwise.

### GetStageOk

`func (o *NodePoolSetting) GetStageOk() (*string, bool)`

GetStageOk returns a tuple with the Stage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStage

`func (o *NodePoolSetting) SetStage(v string)`

SetStage sets Stage field to given value.

### HasStage

`func (o *NodePoolSetting) HasStage() bool`

HasStage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


