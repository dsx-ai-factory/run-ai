# RegistryIntegration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | **time.Time** | The date and time the registry integration was created. | 
**UpdatedAt** | **time.Time** | The date and time the registry integration was updated. | 
**CreatedBy** | **string** | The user who created the registry integration. | 
**UpdatedBy** | **string** | The user who updated the registry integration. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ScopeId** | Pointer to **string** | The id of the cluster, department or project, depending on the scope type. | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**DepartmentId** | Pointer to **string** | The id of the department. | [optional] 
**ProjectId** | Pointer to **string** | The id of the project. | [optional] 
**Id** | **string** | The ID of the registry integration. | 
**Name** | **string** | The name of the registry integration. | 
**Url** | **string** | The type of the registry integration. | 
**Description** | Pointer to **string** | The description of the registry integration. | [optional] 

## Methods

### NewRegistryIntegration

`func NewRegistryIntegration(createdAt time.Time, updatedAt time.Time, createdBy string, updatedBy string, scopeType ScopeType, id string, name string, url string, ) *RegistryIntegration`

NewRegistryIntegration instantiates a new RegistryIntegration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryIntegrationWithDefaults

`func NewRegistryIntegrationWithDefaults() *RegistryIntegration`

NewRegistryIntegrationWithDefaults instantiates a new RegistryIntegration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *RegistryIntegration) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *RegistryIntegration) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *RegistryIntegration) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *RegistryIntegration) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *RegistryIntegration) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *RegistryIntegration) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedBy

`func (o *RegistryIntegration) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *RegistryIntegration) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *RegistryIntegration) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetUpdatedBy

`func (o *RegistryIntegration) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *RegistryIntegration) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *RegistryIntegration) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetScopeType

`func (o *RegistryIntegration) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *RegistryIntegration) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *RegistryIntegration) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *RegistryIntegration) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *RegistryIntegration) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *RegistryIntegration) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *RegistryIntegration) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.

### GetTenantId

`func (o *RegistryIntegration) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *RegistryIntegration) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *RegistryIntegration) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *RegistryIntegration) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterId

`func (o *RegistryIntegration) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *RegistryIntegration) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *RegistryIntegration) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *RegistryIntegration) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetDepartmentId

`func (o *RegistryIntegration) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *RegistryIntegration) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *RegistryIntegration) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *RegistryIntegration) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetProjectId

`func (o *RegistryIntegration) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *RegistryIntegration) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *RegistryIntegration) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *RegistryIntegration) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetId

`func (o *RegistryIntegration) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RegistryIntegration) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RegistryIntegration) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *RegistryIntegration) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RegistryIntegration) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RegistryIntegration) SetName(v string)`

SetName sets Name field to given value.


### GetUrl

`func (o *RegistryIntegration) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *RegistryIntegration) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *RegistryIntegration) SetUrl(v string)`

SetUrl sets Url field to given value.


### GetDescription

`func (o *RegistryIntegration) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RegistryIntegration) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RegistryIntegration) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *RegistryIntegration) HasDescription() bool`

HasDescription returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


