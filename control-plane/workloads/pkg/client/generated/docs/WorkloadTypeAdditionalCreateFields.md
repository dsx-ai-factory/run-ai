# WorkloadTypeAdditionalCreateFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | The unique name of the workload type. This value must exactly match the Kubernetes Kind that represents the workload type. | [optional] 
**Group** | Pointer to **string** | The Kubernetes group associated with the workload resource. | [optional] 

## Methods

### NewWorkloadTypeAdditionalCreateFields

`func NewWorkloadTypeAdditionalCreateFields() *WorkloadTypeAdditionalCreateFields`

NewWorkloadTypeAdditionalCreateFields instantiates a new WorkloadTypeAdditionalCreateFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeAdditionalCreateFieldsWithDefaults

`func NewWorkloadTypeAdditionalCreateFieldsWithDefaults() *WorkloadTypeAdditionalCreateFields`

NewWorkloadTypeAdditionalCreateFieldsWithDefaults instantiates a new WorkloadTypeAdditionalCreateFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadTypeAdditionalCreateFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadTypeAdditionalCreateFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadTypeAdditionalCreateFields) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkloadTypeAdditionalCreateFields) HasName() bool`

HasName returns a boolean if a field has been set.

### GetGroup

`func (o *WorkloadTypeAdditionalCreateFields) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *WorkloadTypeAdditionalCreateFields) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *WorkloadTypeAdditionalCreateFields) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *WorkloadTypeAdditionalCreateFields) HasGroup() bool`

HasGroup returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


