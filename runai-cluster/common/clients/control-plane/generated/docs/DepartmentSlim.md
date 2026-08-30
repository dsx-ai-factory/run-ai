# DepartmentSlim

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | Pointer to [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 

## Methods

### NewDepartmentSlim

`func NewDepartmentSlim() *DepartmentSlim`

NewDepartmentSlim instantiates a new DepartmentSlim object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentSlimWithDefaults

`func NewDepartmentSlimWithDefaults() *DepartmentSlim`

NewDepartmentSlimWithDefaults instantiates a new DepartmentSlim object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *DepartmentSlim) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentSlim) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentSlim) SetResources(v []Resources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentSlim) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetName

`func (o *DepartmentSlim) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentSlim) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentSlim) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DepartmentSlim) HasName() bool`

HasName returns a boolean if a field has been set.

### GetClusterId

`func (o *DepartmentSlim) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DepartmentSlim) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DepartmentSlim) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *DepartmentSlim) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetId

`func (o *DepartmentSlim) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DepartmentSlim) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DepartmentSlim) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DepartmentSlim) HasId() bool`

HasId returns a boolean if a field has been set.

### GetCreatedAt

`func (o *DepartmentSlim) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DepartmentSlim) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DepartmentSlim) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DepartmentSlim) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *DepartmentSlim) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DepartmentSlim) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DepartmentSlim) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *DepartmentSlim) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *DepartmentSlim) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *DepartmentSlim) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *DepartmentSlim) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *DepartmentSlim) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *DepartmentSlim) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *DepartmentSlim) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *DepartmentSlim) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *DepartmentSlim) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


