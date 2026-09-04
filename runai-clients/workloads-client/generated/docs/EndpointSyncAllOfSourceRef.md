# EndpointSyncAllOfSourceRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Kind** | Pointer to **string** | The Kubernetes resource kind that exposes the endpoint (for example, Ingress or Service). | [optional] 
**Name** | Pointer to **string** | The name of the Kubernetes resource exposing the endpoint. | [optional] 

## Methods

### NewEndpointSyncAllOfSourceRef

`func NewEndpointSyncAllOfSourceRef() *EndpointSyncAllOfSourceRef`

NewEndpointSyncAllOfSourceRef instantiates a new EndpointSyncAllOfSourceRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEndpointSyncAllOfSourceRefWithDefaults

`func NewEndpointSyncAllOfSourceRefWithDefaults() *EndpointSyncAllOfSourceRef`

NewEndpointSyncAllOfSourceRefWithDefaults instantiates a new EndpointSyncAllOfSourceRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKind

`func (o *EndpointSyncAllOfSourceRef) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *EndpointSyncAllOfSourceRef) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *EndpointSyncAllOfSourceRef) SetKind(v string)`

SetKind sets Kind field to given value.

### HasKind

`func (o *EndpointSyncAllOfSourceRef) HasKind() bool`

HasKind returns a boolean if a field has been set.

### GetName

`func (o *EndpointSyncAllOfSourceRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EndpointSyncAllOfSourceRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EndpointSyncAllOfSourceRef) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *EndpointSyncAllOfSourceRef) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


