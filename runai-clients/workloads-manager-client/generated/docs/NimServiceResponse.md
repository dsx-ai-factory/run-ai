# NimServiceResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**WorkloadV2Metadata**](WorkloadV2Metadata.md) |  | 
**DesiredPhase** | [**DesiredPhase**](DesiredPhase.md) |  | 
**Spec** | [**NimServiceSpec**](NimServiceSpec.md) |  | 

## Methods

### NewNimServiceResponse

`func NewNimServiceResponse(metadata WorkloadV2Metadata, desiredPhase DesiredPhase, spec NimServiceSpec, ) *NimServiceResponse`

NewNimServiceResponse instantiates a new NimServiceResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceResponseWithDefaults

`func NewNimServiceResponseWithDefaults() *NimServiceResponse`

NewNimServiceResponseWithDefaults instantiates a new NimServiceResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *NimServiceResponse) GetMetadata() WorkloadV2Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *NimServiceResponse) GetMetadataOk() (*WorkloadV2Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *NimServiceResponse) SetMetadata(v WorkloadV2Metadata)`

SetMetadata sets Metadata field to given value.


### GetDesiredPhase

`func (o *NimServiceResponse) GetDesiredPhase() DesiredPhase`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *NimServiceResponse) GetDesiredPhaseOk() (*DesiredPhase, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *NimServiceResponse) SetDesiredPhase(v DesiredPhase)`

SetDesiredPhase sets DesiredPhase field to given value.


### GetSpec

`func (o *NimServiceResponse) GetSpec() NimServiceSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *NimServiceResponse) GetSpecOk() (*NimServiceSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *NimServiceResponse) SetSpec(v NimServiceSpec)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


