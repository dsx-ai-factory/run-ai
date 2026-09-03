# Cluster

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uuid** | Pointer to **string** | The cluster ID. | [optional] 
**CreatedAt** | Pointer to **string** | The cluster&#39;s creation date. | [optional] 
**Domain** | Pointer to **string** | The cluster&#39;s url. | [optional] 
**TenantId** | Pointer to **float32** | The tenant id this cluster belongs to. | [optional] 
**Name** | Pointer to **string** | The cluster&#39;s name. | [optional] 
**Description** | Pointer to **string** | The cluster&#39;s name. | [optional] 
**Version** | Pointer to **string** | The cluster&#39;s  RUN:AI version. | [optional] 
**Connected** | Pointer to **bool** | if the cluster is connected to the control plane. | [optional] 

## Methods

### NewCluster

`func NewCluster() *Cluster`

NewCluster instantiates a new Cluster object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterWithDefaults

`func NewClusterWithDefaults() *Cluster`

NewClusterWithDefaults instantiates a new Cluster object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUuid

`func (o *Cluster) GetUuid() string`

GetUuid returns the Uuid field if non-nil, zero value otherwise.

### GetUuidOk

`func (o *Cluster) GetUuidOk() (*string, bool)`

GetUuidOk returns a tuple with the Uuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUuid

`func (o *Cluster) SetUuid(v string)`

SetUuid sets Uuid field to given value.

### HasUuid

`func (o *Cluster) HasUuid() bool`

HasUuid returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Cluster) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Cluster) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Cluster) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Cluster) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDomain

`func (o *Cluster) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *Cluster) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *Cluster) SetDomain(v string)`

SetDomain sets Domain field to given value.

### HasDomain

`func (o *Cluster) HasDomain() bool`

HasDomain returns a boolean if a field has been set.

### GetTenantId

`func (o *Cluster) GetTenantId() float32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Cluster) GetTenantIdOk() (*float32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Cluster) SetTenantId(v float32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Cluster) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetName

`func (o *Cluster) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Cluster) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Cluster) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Cluster) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *Cluster) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Cluster) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Cluster) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Cluster) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetVersion

`func (o *Cluster) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *Cluster) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *Cluster) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *Cluster) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetConnected

`func (o *Cluster) GetConnected() bool`

GetConnected returns the Connected field if non-nil, zero value otherwise.

### GetConnectedOk

`func (o *Cluster) GetConnectedOk() (*bool, bool)`

GetConnectedOk returns a tuple with the Connected field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnected

`func (o *Cluster) SetConnected(v bool)`

SetConnected sets Connected field to given value.

### HasConnected

`func (o *Cluster) HasConnected() bool`

HasConnected returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


