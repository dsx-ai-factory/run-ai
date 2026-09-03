# ComputeCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetCreationRequest**](AssetCreationRequest.md) |  | 
**Spec** | [**ComputeAssetSpec**](ComputeAssetSpec.md) |  | 

## Methods

### NewComputeCreationRequest

`func NewComputeCreationRequest(meta AssetCreationRequest, spec ComputeAssetSpec, ) *ComputeCreationRequest`

NewComputeCreationRequest instantiates a new ComputeCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComputeCreationRequestWithDefaults

`func NewComputeCreationRequestWithDefaults() *ComputeCreationRequest`

NewComputeCreationRequestWithDefaults instantiates a new ComputeCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *ComputeCreationRequest) GetMeta() AssetCreationRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *ComputeCreationRequest) GetMetaOk() (*AssetCreationRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *ComputeCreationRequest) SetMeta(v AssetCreationRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *ComputeCreationRequest) GetSpec() ComputeAssetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ComputeCreationRequest) GetSpecOk() (*ComputeAssetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ComputeCreationRequest) SetSpec(v ComputeAssetSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


