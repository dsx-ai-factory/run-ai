# PermittedScopes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Tenant** | Pointer to **string** |  | [optional] 
**Clusters** | Pointer to **[]string** |  | [optional] 
**Departments** | Pointer to **[]string** |  | [optional] 
**Projects** | Pointer to **[]string** |  | [optional] 

## Methods

### NewPermittedScopes

`func NewPermittedScopes() *PermittedScopes`

NewPermittedScopes instantiates a new PermittedScopes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPermittedScopesWithDefaults

`func NewPermittedScopesWithDefaults() *PermittedScopes`

NewPermittedScopesWithDefaults instantiates a new PermittedScopes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenant

`func (o *PermittedScopes) GetTenant() string`

GetTenant returns the Tenant field if non-nil, zero value otherwise.

### GetTenantOk

`func (o *PermittedScopes) GetTenantOk() (*string, bool)`

GetTenantOk returns a tuple with the Tenant field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenant

`func (o *PermittedScopes) SetTenant(v string)`

SetTenant sets Tenant field to given value.

### HasTenant

`func (o *PermittedScopes) HasTenant() bool`

HasTenant returns a boolean if a field has been set.

### GetClusters

`func (o *PermittedScopes) GetClusters() []string`

GetClusters returns the Clusters field if non-nil, zero value otherwise.

### GetClustersOk

`func (o *PermittedScopes) GetClustersOk() (*[]string, bool)`

GetClustersOk returns a tuple with the Clusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusters

`func (o *PermittedScopes) SetClusters(v []string)`

SetClusters sets Clusters field to given value.

### HasClusters

`func (o *PermittedScopes) HasClusters() bool`

HasClusters returns a boolean if a field has been set.

### GetDepartments

`func (o *PermittedScopes) GetDepartments() []string`

GetDepartments returns the Departments field if non-nil, zero value otherwise.

### GetDepartmentsOk

`func (o *PermittedScopes) GetDepartmentsOk() (*[]string, bool)`

GetDepartmentsOk returns a tuple with the Departments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartments

`func (o *PermittedScopes) SetDepartments(v []string)`

SetDepartments sets Departments field to given value.

### HasDepartments

`func (o *PermittedScopes) HasDepartments() bool`

HasDepartments returns a boolean if a field has been set.

### GetProjects

`func (o *PermittedScopes) GetProjects() []string`

GetProjects returns the Projects field if non-nil, zero value otherwise.

### GetProjectsOk

`func (o *PermittedScopes) GetProjectsOk() (*[]string, bool)`

GetProjectsOk returns a tuple with the Projects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjects

`func (o *PermittedScopes) SetProjects(v []string)`

SetProjects sets Projects field to given value.

### HasProjects

`func (o *PermittedScopes) HasProjects() bool`

HasProjects returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


