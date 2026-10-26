# WorkloadNetworkStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Services** | Pointer to [**[]WorkloadNetworkStatusServicesInner**](WorkloadNetworkStatusServicesInner.md) | List of services associated with the workload. | [optional] 

## Methods

### NewWorkloadNetworkStatus

`func NewWorkloadNetworkStatus() *WorkloadNetworkStatus`

NewWorkloadNetworkStatus instantiates a new WorkloadNetworkStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadNetworkStatusWithDefaults

`func NewWorkloadNetworkStatusWithDefaults() *WorkloadNetworkStatus`

NewWorkloadNetworkStatusWithDefaults instantiates a new WorkloadNetworkStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServices

`func (o *WorkloadNetworkStatus) GetServices() []WorkloadNetworkStatusServicesInner`

GetServices returns the Services field if non-nil, zero value otherwise.

### GetServicesOk

`func (o *WorkloadNetworkStatus) GetServicesOk() (*[]WorkloadNetworkStatusServicesInner, bool)`

GetServicesOk returns a tuple with the Services field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServices

`func (o *WorkloadNetworkStatus) SetServices(v []WorkloadNetworkStatusServicesInner)`

SetServices sets Services field to given value.

### HasServices

`func (o *WorkloadNetworkStatus) HasServices() bool`

HasServices returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


