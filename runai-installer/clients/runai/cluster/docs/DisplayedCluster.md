# DisplayedCluster

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uuid** | **string** |  | 
**TenantId** | **int32** | The id of the tenant. | 
**Name** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**Domain** | Pointer to **NullableString** |  | [optional] 
**Version** | Pointer to **NullableString** |  | [optional] 
**Status** | Pointer to [**NullableClusterDisplayedStatus**](ClusterDisplayedStatus.md) |  | [optional] 
**UpdatedAt** | Pointer to **NullableTime** |  | [optional] 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**LastLiveness** | Pointer to **NullableTime** |  | [optional] 
**DeleteRequestedAt** | Pointer to **NullableTime** | the timestamp value of when the cluster deletion request was received | [optional] 

## Methods

### NewDisplayedCluster

`func NewDisplayedCluster(uuid string, tenantId int32, name string, createdAt time.Time, ) *DisplayedCluster`

NewDisplayedCluster instantiates a new DisplayedCluster object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDisplayedClusterWithDefaults

`func NewDisplayedClusterWithDefaults() *DisplayedCluster`

NewDisplayedClusterWithDefaults instantiates a new DisplayedCluster object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUuid

`func (o *DisplayedCluster) GetUuid() string`

GetUuid returns the Uuid field if non-nil, zero value otherwise.

### GetUuidOk

`func (o *DisplayedCluster) GetUuidOk() (*string, bool)`

GetUuidOk returns a tuple with the Uuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUuid

`func (o *DisplayedCluster) SetUuid(v string)`

SetUuid sets Uuid field to given value.


### GetTenantId

`func (o *DisplayedCluster) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *DisplayedCluster) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *DisplayedCluster) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetName

`func (o *DisplayedCluster) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DisplayedCluster) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DisplayedCluster) SetName(v string)`

SetName sets Name field to given value.


### GetCreatedAt

`func (o *DisplayedCluster) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DisplayedCluster) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DisplayedCluster) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetDomain

`func (o *DisplayedCluster) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *DisplayedCluster) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *DisplayedCluster) SetDomain(v string)`

SetDomain sets Domain field to given value.

### HasDomain

`func (o *DisplayedCluster) HasDomain() bool`

HasDomain returns a boolean if a field has been set.

### SetDomainNil

`func (o *DisplayedCluster) SetDomainNil(b bool)`

 SetDomainNil sets the value for Domain to be an explicit nil

### UnsetDomain
`func (o *DisplayedCluster) UnsetDomain()`

UnsetDomain ensures that no value is present for Domain, not even an explicit nil
### GetVersion

`func (o *DisplayedCluster) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *DisplayedCluster) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *DisplayedCluster) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *DisplayedCluster) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### SetVersionNil

`func (o *DisplayedCluster) SetVersionNil(b bool)`

 SetVersionNil sets the value for Version to be an explicit nil

### UnsetVersion
`func (o *DisplayedCluster) UnsetVersion()`

UnsetVersion ensures that no value is present for Version, not even an explicit nil
### GetStatus

`func (o *DisplayedCluster) GetStatus() ClusterDisplayedStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DisplayedCluster) GetStatusOk() (*ClusterDisplayedStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DisplayedCluster) SetStatus(v ClusterDisplayedStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DisplayedCluster) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *DisplayedCluster) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *DisplayedCluster) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetUpdatedAt

`func (o *DisplayedCluster) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DisplayedCluster) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DisplayedCluster) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *DisplayedCluster) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### SetUpdatedAtNil

`func (o *DisplayedCluster) SetUpdatedAtNil(b bool)`

 SetUpdatedAtNil sets the value for UpdatedAt to be an explicit nil

### UnsetUpdatedAt
`func (o *DisplayedCluster) UnsetUpdatedAt()`

UnsetUpdatedAt ensures that no value is present for UpdatedAt, not even an explicit nil
### GetDeletedAt

`func (o *DisplayedCluster) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *DisplayedCluster) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *DisplayedCluster) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *DisplayedCluster) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *DisplayedCluster) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *DisplayedCluster) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetLastLiveness

`func (o *DisplayedCluster) GetLastLiveness() time.Time`

GetLastLiveness returns the LastLiveness field if non-nil, zero value otherwise.

### GetLastLivenessOk

`func (o *DisplayedCluster) GetLastLivenessOk() (*time.Time, bool)`

GetLastLivenessOk returns a tuple with the LastLiveness field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastLiveness

`func (o *DisplayedCluster) SetLastLiveness(v time.Time)`

SetLastLiveness sets LastLiveness field to given value.

### HasLastLiveness

`func (o *DisplayedCluster) HasLastLiveness() bool`

HasLastLiveness returns a boolean if a field has been set.

### SetLastLivenessNil

`func (o *DisplayedCluster) SetLastLivenessNil(b bool)`

 SetLastLivenessNil sets the value for LastLiveness to be an explicit nil

### UnsetLastLiveness
`func (o *DisplayedCluster) UnsetLastLiveness()`

UnsetLastLiveness ensures that no value is present for LastLiveness, not even an explicit nil
### GetDeleteRequestedAt

`func (o *DisplayedCluster) GetDeleteRequestedAt() time.Time`

GetDeleteRequestedAt returns the DeleteRequestedAt field if non-nil, zero value otherwise.

### GetDeleteRequestedAtOk

`func (o *DisplayedCluster) GetDeleteRequestedAtOk() (*time.Time, bool)`

GetDeleteRequestedAtOk returns a tuple with the DeleteRequestedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleteRequestedAt

`func (o *DisplayedCluster) SetDeleteRequestedAt(v time.Time)`

SetDeleteRequestedAt sets DeleteRequestedAt field to given value.

### HasDeleteRequestedAt

`func (o *DisplayedCluster) HasDeleteRequestedAt() bool`

HasDeleteRequestedAt returns a boolean if a field has been set.

### SetDeleteRequestedAtNil

`func (o *DisplayedCluster) SetDeleteRequestedAtNil(b bool)`

 SetDeleteRequestedAtNil sets the value for DeleteRequestedAt to be an explicit nil

### UnsetDeleteRequestedAt
`func (o *DisplayedCluster) UnsetDeleteRequestedAt()`

UnsetDeleteRequestedAt ensures that no value is present for DeleteRequestedAt, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


