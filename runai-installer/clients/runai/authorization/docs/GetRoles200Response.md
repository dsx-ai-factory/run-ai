# GetRoles200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Roles** | [**[]Role**](Role.md) |  | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewGetRoles200Response

`func NewGetRoles200Response(roles []Role, ) *GetRoles200Response`

NewGetRoles200Response instantiates a new GetRoles200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetRoles200ResponseWithDefaults

`func NewGetRoles200ResponseWithDefaults() *GetRoles200Response`

NewGetRoles200ResponseWithDefaults instantiates a new GetRoles200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRoles

`func (o *GetRoles200Response) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *GetRoles200Response) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *GetRoles200Response) SetRoles(v []Role)`

SetRoles sets Roles field to given value.


### GetNext

`func (o *GetRoles200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetRoles200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetRoles200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetRoles200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


