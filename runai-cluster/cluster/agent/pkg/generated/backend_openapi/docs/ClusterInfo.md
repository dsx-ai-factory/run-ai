# ClusterInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Domain** | **string** | the cluster domain | 
**Version** | **string** | the cluster version | 
**Dependencies** | Pointer to [**ClusterInfoDependencies**](ClusterInfoDependencies.md) |  | [optional] 

## Methods

### NewClusterInfo

`func NewClusterInfo(domain string, version string, ) *ClusterInfo`

NewClusterInfo instantiates a new ClusterInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterInfoWithDefaults

`func NewClusterInfoWithDefaults() *ClusterInfo`

NewClusterInfoWithDefaults instantiates a new ClusterInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDomain

`func (o *ClusterInfo) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *ClusterInfo) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *ClusterInfo) SetDomain(v string)`

SetDomain sets Domain field to given value.


### GetVersion

`func (o *ClusterInfo) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ClusterInfo) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ClusterInfo) SetVersion(v string)`

SetVersion sets Version field to given value.


### GetDependencies

`func (o *ClusterInfo) GetDependencies() ClusterInfoDependencies`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *ClusterInfo) GetDependenciesOk() (*ClusterInfoDependencies, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *ClusterInfo) SetDependencies(v ClusterInfoDependencies)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *ClusterInfo) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


