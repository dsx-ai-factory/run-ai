# ProjectQuotasV1Inner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **int32** | Project id. | [optional] [readonly] 
**TenantId** | Pointer to **int32** | ID of the tenant where the project is located. | [optional] 
**ClusterUuid** | Pointer to **string** | ID of the cluster where the project is located. | [optional] 
**Name** | Pointer to **string** | Project name. | [optional] 
**DeservedGpus** | Pointer to **float32** | The project&#39;s deserved GPU allocation in case the cluster has those resources. | [optional] 
**AllocatedGpus** | Pointer to **float32** | The project&#39;s allocated GPU. | [optional] 

## Methods

### NewProjectQuotasV1Inner

`func NewProjectQuotasV1Inner() *ProjectQuotasV1Inner`

NewProjectQuotasV1Inner instantiates a new ProjectQuotasV1Inner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectQuotasV1InnerWithDefaults

`func NewProjectQuotasV1InnerWithDefaults() *ProjectQuotasV1Inner`

NewProjectQuotasV1InnerWithDefaults instantiates a new ProjectQuotasV1Inner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectQuotasV1Inner) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectQuotasV1Inner) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectQuotasV1Inner) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *ProjectQuotasV1Inner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetTenantId

`func (o *ProjectQuotasV1Inner) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ProjectQuotasV1Inner) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ProjectQuotasV1Inner) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *ProjectQuotasV1Inner) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *ProjectQuotasV1Inner) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *ProjectQuotasV1Inner) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *ProjectQuotasV1Inner) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *ProjectQuotasV1Inner) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetName

`func (o *ProjectQuotasV1Inner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectQuotasV1Inner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectQuotasV1Inner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProjectQuotasV1Inner) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *ProjectQuotasV1Inner) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *ProjectQuotasV1Inner) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *ProjectQuotasV1Inner) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *ProjectQuotasV1Inner) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllocatedGpus

`func (o *ProjectQuotasV1Inner) GetAllocatedGpus() float32`

GetAllocatedGpus returns the AllocatedGpus field if non-nil, zero value otherwise.

### GetAllocatedGpusOk

`func (o *ProjectQuotasV1Inner) GetAllocatedGpusOk() (*float32, bool)`

GetAllocatedGpusOk returns a tuple with the AllocatedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedGpus

`func (o *ProjectQuotasV1Inner) SetAllocatedGpus(v float32)`

SetAllocatedGpus sets AllocatedGpus field to given value.

### HasAllocatedGpus

`func (o *ProjectQuotasV1Inner) HasAllocatedGpus() bool`

HasAllocatedGpus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


