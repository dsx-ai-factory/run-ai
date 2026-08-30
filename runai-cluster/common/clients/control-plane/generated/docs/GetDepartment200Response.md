# GetDepartment200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | Pointer to [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | [optional] 
**Name** | **string** |  | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Id** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Children** | Pointer to [**[]DepartmentProjectChild**](DepartmentProjectChild.md) |  | [optional] 

## Methods

### NewGetDepartment200Response

`func NewGetDepartment200Response(name string, id string, createdAt time.Time, updatedAt time.Time, ) *GetDepartment200Response`

NewGetDepartment200Response instantiates a new GetDepartment200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetDepartment200ResponseWithDefaults

`func NewGetDepartment200ResponseWithDefaults() *GetDepartment200Response`

NewGetDepartment200ResponseWithDefaults instantiates a new GetDepartment200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *GetDepartment200Response) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *GetDepartment200Response) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *GetDepartment200Response) SetResources(v []Resources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *GetDepartment200Response) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetName

`func (o *GetDepartment200Response) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GetDepartment200Response) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GetDepartment200Response) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *GetDepartment200Response) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *GetDepartment200Response) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *GetDepartment200Response) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *GetDepartment200Response) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetId

`func (o *GetDepartment200Response) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetDepartment200Response) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetDepartment200Response) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *GetDepartment200Response) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetDepartment200Response) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetDepartment200Response) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *GetDepartment200Response) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GetDepartment200Response) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GetDepartment200Response) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedBy

`func (o *GetDepartment200Response) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *GetDepartment200Response) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *GetDepartment200Response) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *GetDepartment200Response) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *GetDepartment200Response) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *GetDepartment200Response) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *GetDepartment200Response) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *GetDepartment200Response) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetChildren

`func (o *GetDepartment200Response) GetChildren() []DepartmentProjectChild`

GetChildren returns the Children field if non-nil, zero value otherwise.

### GetChildrenOk

`func (o *GetDepartment200Response) GetChildrenOk() (*[]DepartmentProjectChild, bool)`

GetChildrenOk returns a tuple with the Children field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildren

`func (o *GetDepartment200Response) SetChildren(v []DepartmentProjectChild)`

SetChildren sets Children field to given value.

### HasChildren

`func (o *GetDepartment200Response) HasChildren() bool`

HasChildren returns a boolean if a field has been set.

### SetChildrenNil

`func (o *GetDepartment200Response) SetChildrenNil(b bool)`

 SetChildrenNil sets the value for Children to be an explicit nil

### UnsetChildren
`func (o *GetDepartment200Response) UnsetChildren()`

UnsetChildren ensures that no value is present for Children, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


