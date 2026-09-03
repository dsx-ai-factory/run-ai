# GetDepartments200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Departments** | [**[]DepartmentForList**](DepartmentForList.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetDepartments200Response

`func NewGetDepartments200Response(departments []DepartmentForList, ) *GetDepartments200Response`

NewGetDepartments200Response instantiates a new GetDepartments200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetDepartments200ResponseWithDefaults

`func NewGetDepartments200ResponseWithDefaults() *GetDepartments200Response`

NewGetDepartments200ResponseWithDefaults instantiates a new GetDepartments200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDepartments

`func (o *GetDepartments200Response) GetDepartments() []DepartmentForList`

GetDepartments returns the Departments field if non-nil, zero value otherwise.

### GetDepartmentsOk

`func (o *GetDepartments200Response) GetDepartmentsOk() (*[]DepartmentForList, bool)`

GetDepartmentsOk returns a tuple with the Departments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartments

`func (o *GetDepartments200Response) SetDepartments(v []DepartmentForList)`

SetDepartments sets Departments field to given value.


### GetNext

`func (o *GetDepartments200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetDepartments200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetDepartments200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetDepartments200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


