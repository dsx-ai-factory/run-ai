# Tenant1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **int32** | The id of the tenant. | 
**MimirId** | Pointer to **string** |  | [optional] 
**Uuid** | **string** |  | 
**Name** | **string** |  | 
**DisplayName** | Pointer to **string** |  | [optional] 
**Smg** | **bool** |  | 
**Status** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**ContractType** | **string** |  | 
**Eula** | Pointer to **string** |  | [optional] 

## Methods

### NewTenant1

`func NewTenant1(id int32, uuid string, name string, smg bool, status string, createdAt time.Time, updatedAt time.Time, contractType string, ) *Tenant1`

NewTenant1 instantiates a new Tenant1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenant1WithDefaults

`func NewTenant1WithDefaults() *Tenant1`

NewTenant1WithDefaults instantiates a new Tenant1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Tenant1) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Tenant1) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Tenant1) SetId(v int32)`

SetId sets Id field to given value.


### GetMimirId

`func (o *Tenant1) GetMimirId() string`

GetMimirId returns the MimirId field if non-nil, zero value otherwise.

### GetMimirIdOk

`func (o *Tenant1) GetMimirIdOk() (*string, bool)`

GetMimirIdOk returns a tuple with the MimirId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMimirId

`func (o *Tenant1) SetMimirId(v string)`

SetMimirId sets MimirId field to given value.

### HasMimirId

`func (o *Tenant1) HasMimirId() bool`

HasMimirId returns a boolean if a field has been set.

### GetUuid

`func (o *Tenant1) GetUuid() string`

GetUuid returns the Uuid field if non-nil, zero value otherwise.

### GetUuidOk

`func (o *Tenant1) GetUuidOk() (*string, bool)`

GetUuidOk returns a tuple with the Uuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUuid

`func (o *Tenant1) SetUuid(v string)`

SetUuid sets Uuid field to given value.


### GetName

`func (o *Tenant1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Tenant1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Tenant1) SetName(v string)`

SetName sets Name field to given value.


### GetDisplayName

`func (o *Tenant1) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *Tenant1) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *Tenant1) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *Tenant1) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetSmg

`func (o *Tenant1) GetSmg() bool`

GetSmg returns the Smg field if non-nil, zero value otherwise.

### GetSmgOk

`func (o *Tenant1) GetSmgOk() (*bool, bool)`

GetSmgOk returns a tuple with the Smg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmg

`func (o *Tenant1) SetSmg(v bool)`

SetSmg sets Smg field to given value.


### GetStatus

`func (o *Tenant1) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Tenant1) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Tenant1) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetCreatedAt

`func (o *Tenant1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Tenant1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Tenant1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *Tenant1) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Tenant1) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Tenant1) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetContractType

`func (o *Tenant1) GetContractType() string`

GetContractType returns the ContractType field if non-nil, zero value otherwise.

### GetContractTypeOk

`func (o *Tenant1) GetContractTypeOk() (*string, bool)`

GetContractTypeOk returns a tuple with the ContractType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractType

`func (o *Tenant1) SetContractType(v string)`

SetContractType sets ContractType field to given value.


### GetEula

`func (o *Tenant1) GetEula() string`

GetEula returns the Eula field if non-nil, zero value otherwise.

### GetEulaOk

`func (o *Tenant1) GetEulaOk() (*string, bool)`

GetEulaOk returns a tuple with the Eula field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEula

`func (o *Tenant1) SetEula(v string)`

SetEula sets Eula field to given value.

### HasEula

`func (o *Tenant1) HasEula() bool`

HasEula returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


