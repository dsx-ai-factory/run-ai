# GetAccessRulesSync200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** |  | [optional] 
**AccessRules** | [**[]AccessRuleSync**](AccessRuleSync.md) |  | 

## Methods

### NewGetAccessRulesSync200Response

`func NewGetAccessRulesSync200Response(accessRules []AccessRuleSync, ) *GetAccessRulesSync200Response`

NewGetAccessRulesSync200Response instantiates a new GetAccessRulesSync200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetAccessRulesSync200ResponseWithDefaults

`func NewGetAccessRulesSync200ResponseWithDefaults() *GetAccessRulesSync200Response`

NewGetAccessRulesSync200ResponseWithDefaults instantiates a new GetAccessRulesSync200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *GetAccessRulesSync200Response) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *GetAccessRulesSync200Response) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *GetAccessRulesSync200Response) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *GetAccessRulesSync200Response) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetAccessRules

`func (o *GetAccessRulesSync200Response) GetAccessRules() []AccessRuleSync`

GetAccessRules returns the AccessRules field if non-nil, zero value otherwise.

### GetAccessRulesOk

`func (o *GetAccessRulesSync200Response) GetAccessRulesOk() (*[]AccessRuleSync, bool)`

GetAccessRulesOk returns a tuple with the AccessRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessRules

`func (o *GetAccessRulesSync200Response) SetAccessRules(v []AccessRuleSync)`

SetAccessRules sets AccessRules field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


