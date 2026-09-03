# GroupedMeasurementResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | specifies what data returned | 
**Labels** | Pointer to **map[string]string** | labels of the metric measurement | [optional] 
**Values** | [**[]GroupedMeasurementResponseValuesInner**](GroupedMeasurementResponseValuesInner.md) |  | 
**Groups** | Pointer to [**[]GroupedMeasurementGroup**](GroupedMeasurementGroup.md) | Group information for this measurement | [optional] 

## Methods

### NewGroupedMeasurementResponse

`func NewGroupedMeasurementResponse(type_ string, values []GroupedMeasurementResponseValuesInner, ) *GroupedMeasurementResponse`

NewGroupedMeasurementResponse instantiates a new GroupedMeasurementResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGroupedMeasurementResponseWithDefaults

`func NewGroupedMeasurementResponseWithDefaults() *GroupedMeasurementResponse`

NewGroupedMeasurementResponseWithDefaults instantiates a new GroupedMeasurementResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *GroupedMeasurementResponse) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *GroupedMeasurementResponse) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *GroupedMeasurementResponse) SetType(v string)`

SetType sets Type field to given value.


### GetLabels

`func (o *GroupedMeasurementResponse) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *GroupedMeasurementResponse) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *GroupedMeasurementResponse) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *GroupedMeasurementResponse) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *GroupedMeasurementResponse) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *GroupedMeasurementResponse) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetValues

`func (o *GroupedMeasurementResponse) GetValues() []GroupedMeasurementResponseValuesInner`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *GroupedMeasurementResponse) GetValuesOk() (*[]GroupedMeasurementResponseValuesInner, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *GroupedMeasurementResponse) SetValues(v []GroupedMeasurementResponseValuesInner)`

SetValues sets Values field to given value.


### SetValuesNil

`func (o *GroupedMeasurementResponse) SetValuesNil(b bool)`

 SetValuesNil sets the value for Values to be an explicit nil

### UnsetValues
`func (o *GroupedMeasurementResponse) UnsetValues()`

UnsetValues ensures that no value is present for Values, not even an explicit nil
### GetGroups

`func (o *GroupedMeasurementResponse) GetGroups() []GroupedMeasurementGroup`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *GroupedMeasurementResponse) GetGroupsOk() (*[]GroupedMeasurementGroup, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *GroupedMeasurementResponse) SetGroups(v []GroupedMeasurementGroup)`

SetGroups sets Groups field to given value.

### HasGroups

`func (o *GroupedMeasurementResponse) HasGroups() bool`

HasGroups returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


