# MeasurementResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | specifies what data returned | 
**Labels** | Pointer to **map[string]string** | labels of the metric measurement | [optional] 
**Values** | [**[]GroupedMeasurementResponseValuesInner**](GroupedMeasurementResponseValuesInner.md) |  | 

## Methods

### NewMeasurementResponse

`func NewMeasurementResponse(type_ string, values []GroupedMeasurementResponseValuesInner, ) *MeasurementResponse`

NewMeasurementResponse instantiates a new MeasurementResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMeasurementResponseWithDefaults

`func NewMeasurementResponseWithDefaults() *MeasurementResponse`

NewMeasurementResponseWithDefaults instantiates a new MeasurementResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *MeasurementResponse) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *MeasurementResponse) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *MeasurementResponse) SetType(v string)`

SetType sets Type field to given value.


### GetLabels

`func (o *MeasurementResponse) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *MeasurementResponse) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *MeasurementResponse) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *MeasurementResponse) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *MeasurementResponse) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *MeasurementResponse) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetValues

`func (o *MeasurementResponse) GetValues() []GroupedMeasurementResponseValuesInner`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *MeasurementResponse) GetValuesOk() (*[]GroupedMeasurementResponseValuesInner, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *MeasurementResponse) SetValues(v []GroupedMeasurementResponseValuesInner)`

SetValues sets Values field to given value.


### SetValuesNil

`func (o *MeasurementResponse) SetValuesNil(b bool)`

 SetValuesNil sets the value for Values to be an explicit nil

### UnsetValues
`func (o *MeasurementResponse) UnsetValues()`

UnsetValues ensures that no value is present for Values, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


