# GetDepartmentNodePoolResources200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolResources** | [**[]DepartmentNodePoolResources**](DepartmentNodePoolResources.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetDepartmentNodePoolResources200Response

`func NewGetDepartmentNodePoolResources200Response(nodePoolResources []DepartmentNodePoolResources, ) *GetDepartmentNodePoolResources200Response`

NewGetDepartmentNodePoolResources200Response instantiates a new GetDepartmentNodePoolResources200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetDepartmentNodePoolResources200ResponseWithDefaults

`func NewGetDepartmentNodePoolResources200ResponseWithDefaults() *GetDepartmentNodePoolResources200Response`

NewGetDepartmentNodePoolResources200ResponseWithDefaults instantiates a new GetDepartmentNodePoolResources200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolResources

`func (o *GetDepartmentNodePoolResources200Response) GetNodePoolResources() []DepartmentNodePoolResources`

GetNodePoolResources returns the NodePoolResources field if non-nil, zero value otherwise.

### GetNodePoolResourcesOk

`func (o *GetDepartmentNodePoolResources200Response) GetNodePoolResourcesOk() (*[]DepartmentNodePoolResources, bool)`

GetNodePoolResourcesOk returns a tuple with the NodePoolResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolResources

`func (o *GetDepartmentNodePoolResources200Response) SetNodePoolResources(v []DepartmentNodePoolResources)`

SetNodePoolResources sets NodePoolResources field to given value.


### GetNext

`func (o *GetDepartmentNodePoolResources200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetDepartmentNodePoolResources200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetDepartmentNodePoolResources200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetDepartmentNodePoolResources200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


