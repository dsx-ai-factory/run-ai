# ClusterInfoSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Domain** | **string** | the cluster domain | 
**Version** | **string** | the cluster version | 
**Dependencies** | Pointer to [**NullableClusterInfoSyncRequestDependencies**](ClusterInfoSyncRequestDependencies.md) |  | [optional] 
**Status** | Pointer to [**NullableClusterInfoSyncRequestStatus**](ClusterInfoSyncRequestStatus.md) |  | [optional] 

## Methods

### NewClusterInfoSyncRequest

`func NewClusterInfoSyncRequest(domain string, version string, ) *ClusterInfoSyncRequest`

NewClusterInfoSyncRequest instantiates a new ClusterInfoSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterInfoSyncRequestWithDefaults

`func NewClusterInfoSyncRequestWithDefaults() *ClusterInfoSyncRequest`

NewClusterInfoSyncRequestWithDefaults instantiates a new ClusterInfoSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDomain

`func (o *ClusterInfoSyncRequest) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *ClusterInfoSyncRequest) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *ClusterInfoSyncRequest) SetDomain(v string)`

SetDomain sets Domain field to given value.


### GetVersion

`func (o *ClusterInfoSyncRequest) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ClusterInfoSyncRequest) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ClusterInfoSyncRequest) SetVersion(v string)`

SetVersion sets Version field to given value.


### GetDependencies

`func (o *ClusterInfoSyncRequest) GetDependencies() ClusterInfoSyncRequestDependencies`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *ClusterInfoSyncRequest) GetDependenciesOk() (*ClusterInfoSyncRequestDependencies, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *ClusterInfoSyncRequest) SetDependencies(v ClusterInfoSyncRequestDependencies)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *ClusterInfoSyncRequest) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.

### SetDependenciesNil

`func (o *ClusterInfoSyncRequest) SetDependenciesNil(b bool)`

 SetDependenciesNil sets the value for Dependencies to be an explicit nil

### UnsetDependencies
`func (o *ClusterInfoSyncRequest) UnsetDependencies()`

UnsetDependencies ensures that no value is present for Dependencies, not even an explicit nil
### GetStatus

`func (o *ClusterInfoSyncRequest) GetStatus() ClusterInfoSyncRequestStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ClusterInfoSyncRequest) GetStatusOk() (*ClusterInfoSyncRequestStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ClusterInfoSyncRequest) SetStatus(v ClusterInfoSyncRequestStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ClusterInfoSyncRequest) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *ClusterInfoSyncRequest) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *ClusterInfoSyncRequest) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


