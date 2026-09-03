# GetAccessRules200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TotalRecords** | **int32** |  | 
**DisplayRecords** | **int32** |  | 
**AccessRules** | [**[]AccessRule**](AccessRule.md) |  | 

## Methods

### NewGetAccessRules200Response

`func NewGetAccessRules200Response(totalRecords int32, displayRecords int32, accessRules []AccessRule, ) *GetAccessRules200Response`

NewGetAccessRules200Response instantiates a new GetAccessRules200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetAccessRules200ResponseWithDefaults

`func NewGetAccessRules200ResponseWithDefaults() *GetAccessRules200Response`

NewGetAccessRules200ResponseWithDefaults instantiates a new GetAccessRules200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTotalRecords

`func (o *GetAccessRules200Response) GetTotalRecords() int32`

GetTotalRecords returns the TotalRecords field if non-nil, zero value otherwise.

### GetTotalRecordsOk

`func (o *GetAccessRules200Response) GetTotalRecordsOk() (*int32, bool)`

GetTotalRecordsOk returns a tuple with the TotalRecords field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRecords

`func (o *GetAccessRules200Response) SetTotalRecords(v int32)`

SetTotalRecords sets TotalRecords field to given value.


### GetDisplayRecords

`func (o *GetAccessRules200Response) GetDisplayRecords() int32`

GetDisplayRecords returns the DisplayRecords field if non-nil, zero value otherwise.

### GetDisplayRecordsOk

`func (o *GetAccessRules200Response) GetDisplayRecordsOk() (*int32, bool)`

GetDisplayRecordsOk returns a tuple with the DisplayRecords field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayRecords

`func (o *GetAccessRules200Response) SetDisplayRecords(v int32)`

SetDisplayRecords sets DisplayRecords field to given value.


### GetAccessRules

`func (o *GetAccessRules200Response) GetAccessRules() []AccessRule`

GetAccessRules returns the AccessRules field if non-nil, zero value otherwise.

### GetAccessRulesOk

`func (o *GetAccessRules200Response) GetAccessRulesOk() (*[]AccessRule, bool)`

GetAccessRulesOk returns a tuple with the AccessRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessRules

`func (o *GetAccessRules200Response) SetAccessRules(v []AccessRule)`

SetAccessRules sets AccessRules field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


