# Tenant

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **int32** | The id of the tenant. | 
**MimirId** | Pointer to **string** |  | [optional] 
**Uuid** | **string** | The universally unique identifier of the tenant. | 
**Name** | **string** | The internal system name of the tenant. | 
**DisplayName** | Pointer to **string** | The name of the tenant as shown to users within the tenant. | [optional] 
**Alias** | Pointer to **string** | An optional alias for the tenant. | [optional] 
**Smg** | **bool** |  | 
**Status** | **string** | The current status of the tenant. Possible values: - Creating - Failed - Updating - Deleted - Deleting - Ready  | 
**Diagnostics** | Pointer to **string** | Most recent provisioning failure message. Only populated by GET /tenants/{id} when status is Failed. | [optional] 
**CreatedAt** | **time.Time** | The date and time when the tenant was created. | 
**UpdatedAt** | **time.Time** | The date and time when the tenant was last updated. | 
**ContractType** | **string** |  | 

## Methods

### NewTenant

`func NewTenant(id int32, uuid string, name string, smg bool, status string, createdAt time.Time, updatedAt time.Time, contractType string, ) *Tenant`

NewTenant instantiates a new Tenant object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantWithDefaults

`func NewTenantWithDefaults() *Tenant`

NewTenantWithDefaults instantiates a new Tenant object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Tenant) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Tenant) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Tenant) SetId(v int32)`

SetId sets Id field to given value.


### GetMimirId

`func (o *Tenant) GetMimirId() string`

GetMimirId returns the MimirId field if non-nil, zero value otherwise.

### GetMimirIdOk

`func (o *Tenant) GetMimirIdOk() (*string, bool)`

GetMimirIdOk returns a tuple with the MimirId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMimirId

`func (o *Tenant) SetMimirId(v string)`

SetMimirId sets MimirId field to given value.

### HasMimirId

`func (o *Tenant) HasMimirId() bool`

HasMimirId returns a boolean if a field has been set.

### GetUuid

`func (o *Tenant) GetUuid() string`

GetUuid returns the Uuid field if non-nil, zero value otherwise.

### GetUuidOk

`func (o *Tenant) GetUuidOk() (*string, bool)`

GetUuidOk returns a tuple with the Uuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUuid

`func (o *Tenant) SetUuid(v string)`

SetUuid sets Uuid field to given value.


### GetName

`func (o *Tenant) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Tenant) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Tenant) SetName(v string)`

SetName sets Name field to given value.


### GetDisplayName

`func (o *Tenant) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *Tenant) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *Tenant) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *Tenant) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetAlias

`func (o *Tenant) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *Tenant) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *Tenant) SetAlias(v string)`

SetAlias sets Alias field to given value.

### HasAlias

`func (o *Tenant) HasAlias() bool`

HasAlias returns a boolean if a field has been set.

### GetSmg

`func (o *Tenant) GetSmg() bool`

GetSmg returns the Smg field if non-nil, zero value otherwise.

### GetSmgOk

`func (o *Tenant) GetSmgOk() (*bool, bool)`

GetSmgOk returns a tuple with the Smg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmg

`func (o *Tenant) SetSmg(v bool)`

SetSmg sets Smg field to given value.


### GetStatus

`func (o *Tenant) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Tenant) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Tenant) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetDiagnostics

`func (o *Tenant) GetDiagnostics() string`

GetDiagnostics returns the Diagnostics field if non-nil, zero value otherwise.

### GetDiagnosticsOk

`func (o *Tenant) GetDiagnosticsOk() (*string, bool)`

GetDiagnosticsOk returns a tuple with the Diagnostics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDiagnostics

`func (o *Tenant) SetDiagnostics(v string)`

SetDiagnostics sets Diagnostics field to given value.

### HasDiagnostics

`func (o *Tenant) HasDiagnostics() bool`

HasDiagnostics returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Tenant) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Tenant) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Tenant) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *Tenant) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Tenant) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Tenant) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetContractType

`func (o *Tenant) GetContractType() string`

GetContractType returns the ContractType field if non-nil, zero value otherwise.

### GetContractTypeOk

`func (o *Tenant) GetContractTypeOk() (*string, bool)`

GetContractTypeOk returns a tuple with the ContractType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractType

`func (o *Tenant) SetContractType(v string)`

SetContractType sets ContractType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


