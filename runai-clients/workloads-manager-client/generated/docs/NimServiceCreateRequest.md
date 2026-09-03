# NimServiceCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**NIMServiceMetadataCreateParams**](NIMServiceMetadataCreateParams.md) |  | 
**Spec** | [**NimServiceSpec**](NimServiceSpec.md) |  | 

## Methods

### NewNimServiceCreateRequest

`func NewNimServiceCreateRequest(metadata NIMServiceMetadataCreateParams, spec NimServiceSpec, ) *NimServiceCreateRequest`

NewNimServiceCreateRequest instantiates a new NimServiceCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceCreateRequestWithDefaults

`func NewNimServiceCreateRequestWithDefaults() *NimServiceCreateRequest`

NewNimServiceCreateRequestWithDefaults instantiates a new NimServiceCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *NimServiceCreateRequest) GetMetadata() NIMServiceMetadataCreateParams`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *NimServiceCreateRequest) GetMetadataOk() (*NIMServiceMetadataCreateParams, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *NimServiceCreateRequest) SetMetadata(v NIMServiceMetadataCreateParams)`

SetMetadata sets Metadata field to given value.


### GetSpec

`func (o *NimServiceCreateRequest) GetSpec() NimServiceSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *NimServiceCreateRequest) GetSpecOk() (*NimServiceSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *NimServiceCreateRequest) SetSpec(v NimServiceSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


