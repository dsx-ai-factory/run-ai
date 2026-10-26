# GetWorkloadsSetting

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Value** | Pointer to [**GetEnableSettingsValue**](GetEnableSettingsValue.md) |  | [optional] 
**DisplayName** | Pointer to **string** |  | [optional] [readonly] 
**Description** | Pointer to **string** |  | [optional] [readonly] 
**Stage** | Pointer to **string** |  | [optional] [readonly] 

## Methods

### NewGetWorkloadsSetting

`func NewGetWorkloadsSetting() *GetWorkloadsSetting`

NewGetWorkloadsSetting instantiates a new GetWorkloadsSetting object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetWorkloadsSettingWithDefaults

`func NewGetWorkloadsSettingWithDefaults() *GetWorkloadsSetting`

NewGetWorkloadsSettingWithDefaults instantiates a new GetWorkloadsSetting object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValue

`func (o *GetWorkloadsSetting) GetValue() GetEnableSettingsValue`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *GetWorkloadsSetting) GetValueOk() (*GetEnableSettingsValue, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *GetWorkloadsSetting) SetValue(v GetEnableSettingsValue)`

SetValue sets Value field to given value.

### HasValue

`func (o *GetWorkloadsSetting) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetDisplayName

`func (o *GetWorkloadsSetting) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *GetWorkloadsSetting) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *GetWorkloadsSetting) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *GetWorkloadsSetting) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *GetWorkloadsSetting) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GetWorkloadsSetting) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GetWorkloadsSetting) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GetWorkloadsSetting) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetStage

`func (o *GetWorkloadsSetting) GetStage() string`

GetStage returns the Stage field if non-nil, zero value otherwise.

### GetStageOk

`func (o *GetWorkloadsSetting) GetStageOk() (*string, bool)`

GetStageOk returns a tuple with the Stage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStage

`func (o *GetWorkloadsSetting) SetStage(v string)`

SetStage sets Stage field to given value.

### HasStage

`func (o *GetWorkloadsSetting) HasStage() bool`

HasStage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


