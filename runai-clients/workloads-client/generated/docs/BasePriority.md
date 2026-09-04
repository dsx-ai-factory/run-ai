# BasePriority

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier of the workload priority | 
**Source** | Pointer to **string** | The source of the priority class. Indicates where the priority definition originated. Currently, the value is either Run:ai or empty. | [optional] 
**Name** | **string** | The name of the workload priority (e.g., low, medium, high) | 
**Value** | **int32** | The scheduling value assigned to this priority | 
**CreatedAt** | Pointer to **time.Time** | The timestamp for when the priority was created | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The timestamp for when the priority was updated | [optional] 

## Methods

### NewBasePriority

`func NewBasePriority(id string, name string, value int32, ) *BasePriority`

NewBasePriority instantiates a new BasePriority object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBasePriorityWithDefaults

`func NewBasePriorityWithDefaults() *BasePriority`

NewBasePriorityWithDefaults instantiates a new BasePriority object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *BasePriority) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *BasePriority) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *BasePriority) SetId(v string)`

SetId sets Id field to given value.


### GetSource

`func (o *BasePriority) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *BasePriority) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *BasePriority) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *BasePriority) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetName

`func (o *BasePriority) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *BasePriority) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *BasePriority) SetName(v string)`

SetName sets Name field to given value.


### GetValue

`func (o *BasePriority) GetValue() int32`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *BasePriority) GetValueOk() (*int32, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *BasePriority) SetValue(v int32)`

SetValue sets Value field to given value.


### GetCreatedAt

`func (o *BasePriority) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *BasePriority) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *BasePriority) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *BasePriority) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *BasePriority) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *BasePriority) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *BasePriority) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *BasePriority) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


