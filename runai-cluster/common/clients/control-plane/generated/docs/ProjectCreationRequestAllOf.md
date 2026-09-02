# ProjectCreationRequestAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RequestedNamespace** | Pointer to **string** | project&#39;s requested namespace | [optional] 
**ParentId** | Pointer to **NullableString** | department parent uuid | [optional] 

## Methods

### NewProjectCreationRequestAllOf

`func NewProjectCreationRequestAllOf() *ProjectCreationRequestAllOf`

NewProjectCreationRequestAllOf instantiates a new ProjectCreationRequestAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectCreationRequestAllOfWithDefaults

`func NewProjectCreationRequestAllOfWithDefaults() *ProjectCreationRequestAllOf`

NewProjectCreationRequestAllOfWithDefaults instantiates a new ProjectCreationRequestAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequestedNamespace

`func (o *ProjectCreationRequestAllOf) GetRequestedNamespace() string`

GetRequestedNamespace returns the RequestedNamespace field if non-nil, zero value otherwise.

### GetRequestedNamespaceOk

`func (o *ProjectCreationRequestAllOf) GetRequestedNamespaceOk() (*string, bool)`

GetRequestedNamespaceOk returns a tuple with the RequestedNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNamespace

`func (o *ProjectCreationRequestAllOf) SetRequestedNamespace(v string)`

SetRequestedNamespace sets RequestedNamespace field to given value.

### HasRequestedNamespace

`func (o *ProjectCreationRequestAllOf) HasRequestedNamespace() bool`

HasRequestedNamespace returns a boolean if a field has been set.

### GetParentId

`func (o *ProjectCreationRequestAllOf) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *ProjectCreationRequestAllOf) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *ProjectCreationRequestAllOf) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *ProjectCreationRequestAllOf) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### SetParentIdNil

`func (o *ProjectCreationRequestAllOf) SetParentIdNil(b bool)`

 SetParentIdNil sets the value for ParentId to be an explicit nil

### UnsetParentId
`func (o *ProjectCreationRequestAllOf) UnsetParentId()`

UnsetParentId ensures that no value is present for ParentId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


