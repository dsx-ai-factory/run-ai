# InternalResourceClaimTemplateWorkloadsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the ResourceClaimTemplate, as referenced by pods. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ScopeId** | **string** | The id of the cluster, department or project, depending on the scope type. | 

## Methods

### NewInternalResourceClaimTemplateWorkloadsRequest

`func NewInternalResourceClaimTemplateWorkloadsRequest(name string, scopeType ScopeType, scopeId string, ) *InternalResourceClaimTemplateWorkloadsRequest`

NewInternalResourceClaimTemplateWorkloadsRequest instantiates a new InternalResourceClaimTemplateWorkloadsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalResourceClaimTemplateWorkloadsRequestWithDefaults

`func NewInternalResourceClaimTemplateWorkloadsRequestWithDefaults() *InternalResourceClaimTemplateWorkloadsRequest`

NewInternalResourceClaimTemplateWorkloadsRequestWithDefaults instantiates a new InternalResourceClaimTemplateWorkloadsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *InternalResourceClaimTemplateWorkloadsRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InternalResourceClaimTemplateWorkloadsRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InternalResourceClaimTemplateWorkloadsRequest) SetName(v string)`

SetName sets Name field to given value.


### GetScopeType

`func (o *InternalResourceClaimTemplateWorkloadsRequest) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *InternalResourceClaimTemplateWorkloadsRequest) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *InternalResourceClaimTemplateWorkloadsRequest) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *InternalResourceClaimTemplateWorkloadsRequest) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *InternalResourceClaimTemplateWorkloadsRequest) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *InternalResourceClaimTemplateWorkloadsRequest) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


