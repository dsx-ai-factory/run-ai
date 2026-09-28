# PrioritiesForStatusPrioritiesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | The unique identifier of the workload priority | [optional] 
**Name** | Pointer to **string** | The name of the workload priority (e.g., Low, Medium High) | [optional] 
**Value** | Pointer to **int32** | The scheduling value assigned to this priority | [optional] 
**Source** | Pointer to **string** |  | [optional] 
**Status** | Pointer to [**PrioritiesForStatusPrioritiesInnerAllOfStatus**](PrioritiesForStatusPrioritiesInnerAllOfStatus.md) |  | [optional] 

## Methods

### NewPrioritiesForStatusPrioritiesInner

`func NewPrioritiesForStatusPrioritiesInner() *PrioritiesForStatusPrioritiesInner`

NewPrioritiesForStatusPrioritiesInner instantiates a new PrioritiesForStatusPrioritiesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPrioritiesForStatusPrioritiesInnerWithDefaults

`func NewPrioritiesForStatusPrioritiesInnerWithDefaults() *PrioritiesForStatusPrioritiesInner`

NewPrioritiesForStatusPrioritiesInnerWithDefaults instantiates a new PrioritiesForStatusPrioritiesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *PrioritiesForStatusPrioritiesInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PrioritiesForStatusPrioritiesInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PrioritiesForStatusPrioritiesInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *PrioritiesForStatusPrioritiesInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *PrioritiesForStatusPrioritiesInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PrioritiesForStatusPrioritiesInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PrioritiesForStatusPrioritiesInner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *PrioritiesForStatusPrioritiesInner) HasName() bool`

HasName returns a boolean if a field has been set.

### GetValue

`func (o *PrioritiesForStatusPrioritiesInner) GetValue() int32`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *PrioritiesForStatusPrioritiesInner) GetValueOk() (*int32, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *PrioritiesForStatusPrioritiesInner) SetValue(v int32)`

SetValue sets Value field to given value.

### HasValue

`func (o *PrioritiesForStatusPrioritiesInner) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetSource

`func (o *PrioritiesForStatusPrioritiesInner) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *PrioritiesForStatusPrioritiesInner) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *PrioritiesForStatusPrioritiesInner) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *PrioritiesForStatusPrioritiesInner) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetStatus

`func (o *PrioritiesForStatusPrioritiesInner) GetStatus() PrioritiesForStatusPrioritiesInnerAllOfStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PrioritiesForStatusPrioritiesInner) GetStatusOk() (*PrioritiesForStatusPrioritiesInnerAllOfStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PrioritiesForStatusPrioritiesInner) SetStatus(v PrioritiesForStatusPrioritiesInnerAllOfStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *PrioritiesForStatusPrioritiesInner) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


