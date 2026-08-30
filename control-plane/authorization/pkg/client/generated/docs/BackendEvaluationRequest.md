# BackendEvaluationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceId** | Pointer to **NullableString** |  | [optional] 
**ResourceType** | Pointer to [**NullableResourceType**](ResourceType.md) |  | [optional] 
**RequestAction** | Pointer to [**NullableAction**](Action.md) |  | [optional] 

## Methods

### NewBackendEvaluationRequest

`func NewBackendEvaluationRequest() *BackendEvaluationRequest`

NewBackendEvaluationRequest instantiates a new BackendEvaluationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBackendEvaluationRequestWithDefaults

`func NewBackendEvaluationRequestWithDefaults() *BackendEvaluationRequest`

NewBackendEvaluationRequestWithDefaults instantiates a new BackendEvaluationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceId

`func (o *BackendEvaluationRequest) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *BackendEvaluationRequest) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *BackendEvaluationRequest) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.

### HasResourceId

`func (o *BackendEvaluationRequest) HasResourceId() bool`

HasResourceId returns a boolean if a field has been set.

### SetResourceIdNil

`func (o *BackendEvaluationRequest) SetResourceIdNil(b bool)`

 SetResourceIdNil sets the value for ResourceId to be an explicit nil

### UnsetResourceId
`func (o *BackendEvaluationRequest) UnsetResourceId()`

UnsetResourceId ensures that no value is present for ResourceId, not even an explicit nil
### GetResourceType

`func (o *BackendEvaluationRequest) GetResourceType() ResourceType`

GetResourceType returns the ResourceType field if non-nil, zero value otherwise.

### GetResourceTypeOk

`func (o *BackendEvaluationRequest) GetResourceTypeOk() (*ResourceType, bool)`

GetResourceTypeOk returns a tuple with the ResourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceType

`func (o *BackendEvaluationRequest) SetResourceType(v ResourceType)`

SetResourceType sets ResourceType field to given value.

### HasResourceType

`func (o *BackendEvaluationRequest) HasResourceType() bool`

HasResourceType returns a boolean if a field has been set.

### SetResourceTypeNil

`func (o *BackendEvaluationRequest) SetResourceTypeNil(b bool)`

 SetResourceTypeNil sets the value for ResourceType to be an explicit nil

### UnsetResourceType
`func (o *BackendEvaluationRequest) UnsetResourceType()`

UnsetResourceType ensures that no value is present for ResourceType, not even an explicit nil
### GetRequestAction

`func (o *BackendEvaluationRequest) GetRequestAction() Action`

GetRequestAction returns the RequestAction field if non-nil, zero value otherwise.

### GetRequestActionOk

`func (o *BackendEvaluationRequest) GetRequestActionOk() (*Action, bool)`

GetRequestActionOk returns a tuple with the RequestAction field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestAction

`func (o *BackendEvaluationRequest) SetRequestAction(v Action)`

SetRequestAction sets RequestAction field to given value.

### HasRequestAction

`func (o *BackendEvaluationRequest) HasRequestAction() bool`

HasRequestAction returns a boolean if a field has been set.

### SetRequestActionNil

`func (o *BackendEvaluationRequest) SetRequestActionNil(b bool)`

 SetRequestActionNil sets the value for RequestAction to be an explicit nil

### UnsetRequestAction
`func (o *BackendEvaluationRequest) UnsetRequestAction()`

UnsetRequestAction ensures that no value is present for RequestAction, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


