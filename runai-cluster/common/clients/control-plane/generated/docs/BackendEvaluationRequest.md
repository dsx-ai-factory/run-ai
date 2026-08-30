# BackendEvaluationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceId** | Pointer to **NullableString** |  | [optional] 
**ResourceType** | [**ResourceType**](ResourceType.md) |  | 
**RequestAction** | [**Action**](Action.md) |  | 
**ExpectedToAllow** | Pointer to **NullableBool** |  | [optional] 

## Methods

### NewBackendEvaluationRequest

`func NewBackendEvaluationRequest(resourceType ResourceType, requestAction Action, ) *BackendEvaluationRequest`

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


### GetExpectedToAllow

`func (o *BackendEvaluationRequest) GetExpectedToAllow() bool`

GetExpectedToAllow returns the ExpectedToAllow field if non-nil, zero value otherwise.

### GetExpectedToAllowOk

`func (o *BackendEvaluationRequest) GetExpectedToAllowOk() (*bool, bool)`

GetExpectedToAllowOk returns a tuple with the ExpectedToAllow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpectedToAllow

`func (o *BackendEvaluationRequest) SetExpectedToAllow(v bool)`

SetExpectedToAllow sets ExpectedToAllow field to given value.

### HasExpectedToAllow

`func (o *BackendEvaluationRequest) HasExpectedToAllow() bool`

HasExpectedToAllow returns a boolean if a field has been set.

### SetExpectedToAllowNil

`func (o *BackendEvaluationRequest) SetExpectedToAllowNil(b bool)`

 SetExpectedToAllowNil sets the value for ExpectedToAllow to be an explicit nil

### UnsetExpectedToAllow
`func (o *BackendEvaluationRequest) UnsetExpectedToAllow()`

UnsetExpectedToAllow ensures that no value is present for ExpectedToAllow, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


