# ClientCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | [**ClientType**](ClientType.md) |  | 
**ClientId** | **string** |  | 
**TenantId** | Pointer to **string** |  | [optional] 
**Attributes** | Pointer to [**ClientAttributeCreationRequest**](ClientAttributeCreationRequest.md) |  | [optional] 

## Methods

### NewClientCreationRequest

`func NewClientCreationRequest(type_ ClientType, clientId string, ) *ClientCreationRequest`

NewClientCreationRequest instantiates a new ClientCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClientCreationRequestWithDefaults

`func NewClientCreationRequestWithDefaults() *ClientCreationRequest`

NewClientCreationRequestWithDefaults instantiates a new ClientCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *ClientCreationRequest) GetType() ClientType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ClientCreationRequest) GetTypeOk() (*ClientType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ClientCreationRequest) SetType(v ClientType)`

SetType sets Type field to given value.


### GetClientId

`func (o *ClientCreationRequest) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *ClientCreationRequest) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *ClientCreationRequest) SetClientId(v string)`

SetClientId sets ClientId field to given value.


### GetTenantId

`func (o *ClientCreationRequest) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ClientCreationRequest) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ClientCreationRequest) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *ClientCreationRequest) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetAttributes

`func (o *ClientCreationRequest) GetAttributes() ClientAttributeCreationRequest`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *ClientCreationRequest) GetAttributesOk() (*ClientAttributeCreationRequest, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *ClientCreationRequest) SetAttributes(v ClientAttributeCreationRequest)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *ClientCreationRequest) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


