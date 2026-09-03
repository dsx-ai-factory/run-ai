# GetDepartments200ResponseDepartmentsInner

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

## Methods

### NewGetDepartments200ResponseDepartmentsInner

`func NewGetDepartments200ResponseDepartmentsInner(name string, id string, createdAt time.Time, updatedAt time.Time, ) *GetDepartments200ResponseDepartmentsInner`

NewGetDepartments200ResponseDepartmentsInner instantiates a new GetDepartments200ResponseDepartmentsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetDepartments200ResponseDepartmentsInnerWithDefaults

`func NewGetDepartments200ResponseDepartmentsInnerWithDefaults() *GetDepartments200ResponseDepartmentsInner`

NewGetDepartments200ResponseDepartmentsInnerWithDefaults instantiates a new GetDepartments200ResponseDepartmentsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *GetDepartments200ResponseDepartmentsInner) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *GetDepartments200ResponseDepartmentsInner) SetResources(v []Resources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *GetDepartments200ResponseDepartmentsInner) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetName

`func (o *GetDepartments200ResponseDepartmentsInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GetDepartments200ResponseDepartmentsInner) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *GetDepartments200ResponseDepartmentsInner) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *GetDepartments200ResponseDepartmentsInner) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *GetDepartments200ResponseDepartmentsInner) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetId

`func (o *GetDepartments200ResponseDepartmentsInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetDepartments200ResponseDepartmentsInner) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *GetDepartments200ResponseDepartmentsInner) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetDepartments200ResponseDepartmentsInner) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *GetDepartments200ResponseDepartmentsInner) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GetDepartments200ResponseDepartmentsInner) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedBy

`func (o *GetDepartments200ResponseDepartmentsInner) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *GetDepartments200ResponseDepartmentsInner) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *GetDepartments200ResponseDepartmentsInner) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *GetDepartments200ResponseDepartmentsInner) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *GetDepartments200ResponseDepartmentsInner) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *GetDepartments200ResponseDepartmentsInner) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *GetDepartments200ResponseDepartmentsInner) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


