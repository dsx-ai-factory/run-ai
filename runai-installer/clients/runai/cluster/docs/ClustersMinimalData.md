# ClustersMinimalData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uuid** | **string** |  | 
**Name** | **string** |  | 
**TenantId** | **int32** | The id of the tenant. | 
**Version** | Pointer to **NullableString** |  | [optional] 
**Domain** | Pointer to **NullableString** |  | [optional] 
**Platform** | Pointer to [**NullableClusterReportedStatusPlatform**](ClusterReportedStatusPlatform.md) |  | [optional] 
**State** | Pointer to [**ClusterState**](ClusterState.md) |  | [optional] 
**DependenciesStatus** | Pointer to [**ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] 

## Methods

### NewClustersMinimalData

`func NewClustersMinimalData(uuid string, name string, tenantId int32, ) *ClustersMinimalData`

NewClustersMinimalData instantiates a new ClustersMinimalData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClustersMinimalDataWithDefaults

`func NewClustersMinimalDataWithDefaults() *ClustersMinimalData`

NewClustersMinimalDataWithDefaults instantiates a new ClustersMinimalData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUuid

`func (o *ClustersMinimalData) GetUuid() string`

GetUuid returns the Uuid field if non-nil, zero value otherwise.

### GetUuidOk

`func (o *ClustersMinimalData) GetUuidOk() (*string, bool)`

GetUuidOk returns a tuple with the Uuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUuid

`func (o *ClustersMinimalData) SetUuid(v string)`

SetUuid sets Uuid field to given value.


### GetName

`func (o *ClustersMinimalData) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ClustersMinimalData) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ClustersMinimalData) SetName(v string)`

SetName sets Name field to given value.


### GetTenantId

`func (o *ClustersMinimalData) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ClustersMinimalData) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ClustersMinimalData) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetVersion

`func (o *ClustersMinimalData) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ClustersMinimalData) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ClustersMinimalData) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *ClustersMinimalData) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### SetVersionNil

`func (o *ClustersMinimalData) SetVersionNil(b bool)`

 SetVersionNil sets the value for Version to be an explicit nil

### UnsetVersion
`func (o *ClustersMinimalData) UnsetVersion()`

UnsetVersion ensures that no value is present for Version, not even an explicit nil
### GetDomain

`func (o *ClustersMinimalData) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *ClustersMinimalData) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *ClustersMinimalData) SetDomain(v string)`

SetDomain sets Domain field to given value.

### HasDomain

`func (o *ClustersMinimalData) HasDomain() bool`

HasDomain returns a boolean if a field has been set.

### SetDomainNil

`func (o *ClustersMinimalData) SetDomainNil(b bool)`

 SetDomainNil sets the value for Domain to be an explicit nil

### UnsetDomain
`func (o *ClustersMinimalData) UnsetDomain()`

UnsetDomain ensures that no value is present for Domain, not even an explicit nil
### GetPlatform

`func (o *ClustersMinimalData) GetPlatform() ClusterReportedStatusPlatform`

GetPlatform returns the Platform field if non-nil, zero value otherwise.

### GetPlatformOk

`func (o *ClustersMinimalData) GetPlatformOk() (*ClusterReportedStatusPlatform, bool)`

GetPlatformOk returns a tuple with the Platform field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlatform

`func (o *ClustersMinimalData) SetPlatform(v ClusterReportedStatusPlatform)`

SetPlatform sets Platform field to given value.

### HasPlatform

`func (o *ClustersMinimalData) HasPlatform() bool`

HasPlatform returns a boolean if a field has been set.

### SetPlatformNil

`func (o *ClustersMinimalData) SetPlatformNil(b bool)`

 SetPlatformNil sets the value for Platform to be an explicit nil

### UnsetPlatform
`func (o *ClustersMinimalData) UnsetPlatform()`

UnsetPlatform ensures that no value is present for Platform, not even an explicit nil
### GetState

`func (o *ClustersMinimalData) GetState() ClusterState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ClustersMinimalData) GetStateOk() (*ClusterState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ClustersMinimalData) SetState(v ClusterState)`

SetState sets State field to given value.

### HasState

`func (o *ClustersMinimalData) HasState() bool`

HasState returns a boolean if a field has been set.

### GetDependenciesStatus

`func (o *ClustersMinimalData) GetDependenciesStatus() ClusterDependenciesStatus`

GetDependenciesStatus returns the DependenciesStatus field if non-nil, zero value otherwise.

### GetDependenciesStatusOk

`func (o *ClustersMinimalData) GetDependenciesStatusOk() (*ClusterDependenciesStatus, bool)`

GetDependenciesStatusOk returns a tuple with the DependenciesStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependenciesStatus

`func (o *ClustersMinimalData) SetDependenciesStatus(v ClusterDependenciesStatus)`

SetDependenciesStatus sets DependenciesStatus field to given value.

### HasDependenciesStatus

`func (o *ClustersMinimalData) HasDependenciesStatus() bool`

HasDependenciesStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


