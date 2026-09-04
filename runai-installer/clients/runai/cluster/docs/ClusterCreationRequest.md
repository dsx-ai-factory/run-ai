# ClusterCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Domain** | Pointer to **string** |  | [optional] 
**Version** | Pointer to **string** |  | [optional] 
**TenantId** | Pointer to **int** |  | [optional] 
**DefaultNodepoolConfig** | Pointer to [**DefaultNodepoolConfig**](DefaultNodepoolConfig.md) |  | [optional] 

## Methods

### NewClusterCreationRequest

`func NewClusterCreationRequest(name string, ) *ClusterCreationRequest`

NewClusterCreationRequest instantiates a new ClusterCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterCreationRequestWithDefaults

`func NewClusterCreationRequestWithDefaults() *ClusterCreationRequest`

NewClusterCreationRequestWithDefaults instantiates a new ClusterCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ClusterCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ClusterCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ClusterCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDomain

`func (o *ClusterCreationRequest) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *ClusterCreationRequest) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *ClusterCreationRequest) SetDomain(v string)`

SetDomain sets Domain field to given value.

### HasDomain

`func (o *ClusterCreationRequest) HasDomain() bool`

HasDomain returns a boolean if a field has been set.

### GetVersion

`func (o *ClusterCreationRequest) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ClusterCreationRequest) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ClusterCreationRequest) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *ClusterCreationRequest) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetTenantId

`func (o *ClusterCreationRequest) GetTenantId() int`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ClusterCreationRequest) GetTenantIdOk() (*int, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ClusterCreationRequest) SetTenantId(v int)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *ClusterCreationRequest) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetDefaultNodepoolConfig

`func (o *ClusterCreationRequest) GetDefaultNodepoolConfig() DefaultNodepoolConfig`

GetDefaultNodepoolConfig returns the DefaultNodepoolConfig field if non-nil, zero value otherwise.

### GetDefaultNodepoolConfigOk

`func (o *ClusterCreationRequest) GetDefaultNodepoolConfigOk() (*DefaultNodepoolConfig, bool)`

GetDefaultNodepoolConfigOk returns a tuple with the DefaultNodepoolConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodepoolConfig

`func (o *ClusterCreationRequest) SetDefaultNodepoolConfig(v DefaultNodepoolConfig)`

SetDefaultNodepoolConfig sets DefaultNodepoolConfig field to given value.

### HasDefaultNodepoolConfig

`func (o *ClusterCreationRequest) HasDefaultNodepoolConfig() bool`

HasDefaultNodepoolConfig returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


