# DependencyStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Available** | **bool** | whether the dependency is available | 
**Reason** | Pointer to **string** | the reason for the dependency status | [optional] 
**Components** | Pointer to [**map[string]DependencyStatus**](DependencyStatus.md) |  | [optional] 

## Methods

### NewDependencyStatus

`func NewDependencyStatus(available bool, ) *DependencyStatus`

NewDependencyStatus instantiates a new DependencyStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDependencyStatusWithDefaults

`func NewDependencyStatusWithDefaults() *DependencyStatus`

NewDependencyStatusWithDefaults instantiates a new DependencyStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvailable

`func (o *DependencyStatus) GetAvailable() bool`

GetAvailable returns the Available field if non-nil, zero value otherwise.

### GetAvailableOk

`func (o *DependencyStatus) GetAvailableOk() (*bool, bool)`

GetAvailableOk returns a tuple with the Available field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvailable

`func (o *DependencyStatus) SetAvailable(v bool)`

SetAvailable sets Available field to given value.


### GetReason

`func (o *DependencyStatus) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *DependencyStatus) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *DependencyStatus) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *DependencyStatus) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetComponents

`func (o *DependencyStatus) GetComponents() map[string]DependencyStatus`

GetComponents returns the Components field if non-nil, zero value otherwise.

### GetComponentsOk

`func (o *DependencyStatus) GetComponentsOk() (*map[string]DependencyStatus, bool)`

GetComponentsOk returns a tuple with the Components field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComponents

`func (o *DependencyStatus) SetComponents(v map[string]DependencyStatus)`

SetComponents sets Components field to given value.

### HasComponents

`func (o *DependencyStatus) HasComponents() bool`

HasComponents returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


