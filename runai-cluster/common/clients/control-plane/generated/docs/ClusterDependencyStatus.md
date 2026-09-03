# ClusterDependencyStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Available** | **bool** | whether the dependency is available | 
**Reason** | Pointer to **string** | the reason for the dependency status | [optional] 
**Components** | Pointer to [**map[string]ClusterDependencyStatus**](ClusterDependencyStatus.md) |  | [optional] 

## Methods

### NewClusterDependencyStatus

`func NewClusterDependencyStatus(available bool, ) *ClusterDependencyStatus`

NewClusterDependencyStatus instantiates a new ClusterDependencyStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterDependencyStatusWithDefaults

`func NewClusterDependencyStatusWithDefaults() *ClusterDependencyStatus`

NewClusterDependencyStatusWithDefaults instantiates a new ClusterDependencyStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvailable

`func (o *ClusterDependencyStatus) GetAvailable() bool`

GetAvailable returns the Available field if non-nil, zero value otherwise.

### GetAvailableOk

`func (o *ClusterDependencyStatus) GetAvailableOk() (*bool, bool)`

GetAvailableOk returns a tuple with the Available field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvailable

`func (o *ClusterDependencyStatus) SetAvailable(v bool)`

SetAvailable sets Available field to given value.


### GetReason

`func (o *ClusterDependencyStatus) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *ClusterDependencyStatus) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *ClusterDependencyStatus) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *ClusterDependencyStatus) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetComponents

`func (o *ClusterDependencyStatus) GetComponents() map[string]ClusterDependencyStatus`

GetComponents returns the Components field if non-nil, zero value otherwise.

### GetComponentsOk

`func (o *ClusterDependencyStatus) GetComponentsOk() (*map[string]ClusterDependencyStatus, bool)`

GetComponentsOk returns a tuple with the Components field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComponents

`func (o *ClusterDependencyStatus) SetComponents(v map[string]ClusterDependencyStatus)`

SetComponents sets Components field to given value.

### HasComponents

`func (o *ClusterDependencyStatus) HasComponents() bool`

HasComponents returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


