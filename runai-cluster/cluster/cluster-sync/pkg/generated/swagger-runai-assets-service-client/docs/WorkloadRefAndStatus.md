# WorkloadRefAndStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the workload. | 
**Name** | **string** | The name of the workspace or training | 
**Status** | Pointer to **string** | The status of the workspace or training. | [optional] 

## Methods

### NewWorkloadRefAndStatus

`func NewWorkloadRefAndStatus(id string, name string, ) *WorkloadRefAndStatus`

NewWorkloadRefAndStatus instantiates a new WorkloadRefAndStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadRefAndStatusWithDefaults

`func NewWorkloadRefAndStatusWithDefaults() *WorkloadRefAndStatus`

NewWorkloadRefAndStatusWithDefaults instantiates a new WorkloadRefAndStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadRefAndStatus) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadRefAndStatus) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadRefAndStatus) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *WorkloadRefAndStatus) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadRefAndStatus) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadRefAndStatus) SetName(v string)`

SetName sets Name field to given value.


### GetStatus

`func (o *WorkloadRefAndStatus) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkloadRefAndStatus) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkloadRefAndStatus) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *WorkloadRefAndStatus) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


