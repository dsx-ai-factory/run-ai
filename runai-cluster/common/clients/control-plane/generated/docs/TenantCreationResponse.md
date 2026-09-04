# TenantCreationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Tenant** | [**Tenant1**](Tenant1.md) |  | 
**AdditionalData** | Pointer to **map[string]interface{}** | additional data for tenant creation | [optional] 

## Methods

### NewTenantCreationResponse

`func NewTenantCreationResponse(tenant Tenant1, ) *TenantCreationResponse`

NewTenantCreationResponse instantiates a new TenantCreationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantCreationResponseWithDefaults

`func NewTenantCreationResponseWithDefaults() *TenantCreationResponse`

NewTenantCreationResponseWithDefaults instantiates a new TenantCreationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenant

`func (o *TenantCreationResponse) GetTenant() Tenant1`

GetTenant returns the Tenant field if non-nil, zero value otherwise.

### GetTenantOk

`func (o *TenantCreationResponse) GetTenantOk() (*Tenant1, bool)`

GetTenantOk returns a tuple with the Tenant field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenant

`func (o *TenantCreationResponse) SetTenant(v Tenant1)`

SetTenant sets Tenant field to given value.


### GetAdditionalData

`func (o *TenantCreationResponse) GetAdditionalData() map[string]interface{}`

GetAdditionalData returns the AdditionalData field if non-nil, zero value otherwise.

### GetAdditionalDataOk

`func (o *TenantCreationResponse) GetAdditionalDataOk() (*map[string]interface{}, bool)`

GetAdditionalDataOk returns a tuple with the AdditionalData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalData

`func (o *TenantCreationResponse) SetAdditionalData(v map[string]interface{})`

SetAdditionalData sets AdditionalData field to given value.

### HasAdditionalData

`func (o *TenantCreationResponse) HasAdditionalData() bool`

HasAdditionalData returns a boolean if a field has been set.

### SetAdditionalDataNil

`func (o *TenantCreationResponse) SetAdditionalDataNil(b bool)`

 SetAdditionalDataNil sets the value for AdditionalData to be an explicit nil

### UnsetAdditionalData
`func (o *TenantCreationResponse) UnsetAdditionalData()`

UnsetAdditionalData ensures that no value is present for AdditionalData, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


