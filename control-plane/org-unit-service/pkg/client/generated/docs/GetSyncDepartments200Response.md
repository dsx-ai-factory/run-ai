# GetSyncDepartments200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Departments** | [**[]DepartmentSync**](DepartmentSync.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetSyncDepartments200Response

`func NewGetSyncDepartments200Response(departments []DepartmentSync, ) *GetSyncDepartments200Response`

NewGetSyncDepartments200Response instantiates a new GetSyncDepartments200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetSyncDepartments200ResponseWithDefaults

`func NewGetSyncDepartments200ResponseWithDefaults() *GetSyncDepartments200Response`

NewGetSyncDepartments200ResponseWithDefaults instantiates a new GetSyncDepartments200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDepartments

`func (o *GetSyncDepartments200Response) GetDepartments() []DepartmentSync`

GetDepartments returns the Departments field if non-nil, zero value otherwise.

### GetDepartmentsOk

`func (o *GetSyncDepartments200Response) GetDepartmentsOk() (*[]DepartmentSync, bool)`

GetDepartmentsOk returns a tuple with the Departments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartments

`func (o *GetSyncDepartments200Response) SetDepartments(v []DepartmentSync)`

SetDepartments sets Departments field to given value.


### GetNext

`func (o *GetSyncDepartments200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetSyncDepartments200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetSyncDepartments200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetSyncDepartments200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


