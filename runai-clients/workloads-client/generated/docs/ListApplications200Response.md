# ListApplications200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**Applications** | [**[]AIApplication**](AIApplication.md) |  | 

## Methods

### NewListApplications200Response

`func NewListApplications200Response(applications []AIApplication, ) *ListApplications200Response`

NewListApplications200Response instantiates a new ListApplications200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListApplications200ResponseWithDefaults

`func NewListApplications200ResponseWithDefaults() *ListApplications200Response`

NewListApplications200ResponseWithDefaults instantiates a new ListApplications200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *ListApplications200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *ListApplications200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *ListApplications200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *ListApplications200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetApplications

`func (o *ListApplications200Response) GetApplications() []AIApplication`

GetApplications returns the Applications field if non-nil, zero value otherwise.

### GetApplicationsOk

`func (o *ListApplications200Response) GetApplicationsOk() (*[]AIApplication, bool)`

GetApplicationsOk returns a tuple with the Applications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplications

`func (o *ListApplications200Response) SetApplications(v []AIApplication)`

SetApplications sets Applications field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


