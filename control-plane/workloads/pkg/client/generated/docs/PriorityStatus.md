# PriorityStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | The unique identifier of the cluster | [optional] 
**Phase** | Pointer to **string** | Current phase of the priority in the specified cluster | [optional] 

## Methods

### NewPriorityStatus

`func NewPriorityStatus() *PriorityStatus`

NewPriorityStatus instantiates a new PriorityStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPriorityStatusWithDefaults

`func NewPriorityStatusWithDefaults() *PriorityStatus`

NewPriorityStatusWithDefaults instantiates a new PriorityStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *PriorityStatus) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PriorityStatus) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PriorityStatus) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *PriorityStatus) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetPhase

`func (o *PriorityStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PriorityStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PriorityStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PriorityStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


