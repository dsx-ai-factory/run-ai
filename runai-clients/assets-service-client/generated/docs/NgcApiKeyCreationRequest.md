# NgcApiKeyCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetCreationRequest**](AssetCreationRequest.md) |  | 
**Spec** | [**NgcApiKeyCreationSpec**](NgcApiKeyCreationSpec.md) |  | 

## Methods

### NewNgcApiKeyCreationRequest

`func NewNgcApiKeyCreationRequest(meta AssetCreationRequest, spec NgcApiKeyCreationSpec, ) *NgcApiKeyCreationRequest`

NewNgcApiKeyCreationRequest instantiates a new NgcApiKeyCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNgcApiKeyCreationRequestWithDefaults

`func NewNgcApiKeyCreationRequestWithDefaults() *NgcApiKeyCreationRequest`

NewNgcApiKeyCreationRequestWithDefaults instantiates a new NgcApiKeyCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *NgcApiKeyCreationRequest) GetMeta() AssetCreationRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *NgcApiKeyCreationRequest) GetMetaOk() (*AssetCreationRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *NgcApiKeyCreationRequest) SetMeta(v AssetCreationRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *NgcApiKeyCreationRequest) GetSpec() NgcApiKeyCreationSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *NgcApiKeyCreationRequest) GetSpecOk() (*NgcApiKeyCreationSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *NgcApiKeyCreationRequest) SetSpec(v NgcApiKeyCreationSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


