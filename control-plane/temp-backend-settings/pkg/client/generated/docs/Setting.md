# Setting

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | **string** |  | 
**Source** | **string** |  | 
**Category** | **string** |  | 
**Label** | **string** |  | 
**Description** | **string** |  | 
**Stage** | Pointer to **string** |  | [optional] 
**Type** | **string** |  | 
**Value** | Pointer to **interface{}** |  | [optional] 
**Active** | **bool** |  | 

## Methods

### NewSetting

`func NewSetting(key string, source string, category string, label string, description string, type_ string, active bool, ) *Setting`

NewSetting instantiates a new Setting object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSettingWithDefaults

`func NewSettingWithDefaults() *Setting`

NewSettingWithDefaults instantiates a new Setting object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *Setting) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *Setting) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *Setting) SetKey(v string)`

SetKey sets Key field to given value.


### GetSource

`func (o *Setting) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *Setting) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *Setting) SetSource(v string)`

SetSource sets Source field to given value.


### GetCategory

`func (o *Setting) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *Setting) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *Setting) SetCategory(v string)`

SetCategory sets Category field to given value.


### GetLabel

`func (o *Setting) GetLabel() string`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *Setting) GetLabelOk() (*string, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *Setting) SetLabel(v string)`

SetLabel sets Label field to given value.


### GetDescription

`func (o *Setting) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Setting) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Setting) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetStage

`func (o *Setting) GetStage() string`

GetStage returns the Stage field if non-nil, zero value otherwise.

### GetStageOk

`func (o *Setting) GetStageOk() (*string, bool)`

GetStageOk returns a tuple with the Stage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStage

`func (o *Setting) SetStage(v string)`

SetStage sets Stage field to given value.

### HasStage

`func (o *Setting) HasStage() bool`

HasStage returns a boolean if a field has been set.

### GetType

`func (o *Setting) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Setting) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Setting) SetType(v string)`

SetType sets Type field to given value.


### GetValue

`func (o *Setting) GetValue() interface{}`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *Setting) GetValueOk() (*interface{}, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *Setting) SetValue(v interface{})`

SetValue sets Value field to given value.

### HasValue

`func (o *Setting) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *Setting) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *Setting) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetActive

`func (o *Setting) GetActive() bool`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *Setting) GetActiveOk() (*bool, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *Setting) SetActive(v bool)`

SetActive sets Active field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


