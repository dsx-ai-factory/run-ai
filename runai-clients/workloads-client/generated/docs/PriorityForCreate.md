# PriorityForCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload priority (e.g., Low, Medium High) | 
**Value** | **int32** | The scheduling value assigned to this priority | 

## Methods

### NewPriorityForCreate

`func NewPriorityForCreate(name string, value int32, ) *PriorityForCreate`

NewPriorityForCreate instantiates a new PriorityForCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPriorityForCreateWithDefaults

`func NewPriorityForCreateWithDefaults() *PriorityForCreate`

NewPriorityForCreateWithDefaults instantiates a new PriorityForCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PriorityForCreate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PriorityForCreate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PriorityForCreate) SetName(v string)`

SetName sets Name field to given value.


### GetValue

`func (o *PriorityForCreate) GetValue() int32`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *PriorityForCreate) GetValueOk() (*int32, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *PriorityForCreate) SetValue(v int32)`

SetValue sets Value field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


