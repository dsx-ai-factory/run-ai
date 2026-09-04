# WorkloadV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**WorkloadV2Metadata**](WorkloadV2Metadata.md) |  | 
**DesiredPhase** | [**DesiredPhase**](DesiredPhase.md) |  | 
**WorkloadManifest** | **map[string]interface{}** | The full Kubernetes manifest for the workload in JSON form. Mutually exclusive with manifestEncodedBase64. | 
**RawManifest** | **map[string]interface{}** | The full Kubernetes manifest for the workload in JSON form. Mutually exclusive with manifestEncodedBase64. | 

## Methods

### NewWorkloadV2

`func NewWorkloadV2(metadata WorkloadV2Metadata, desiredPhase DesiredPhase, workloadManifest map[string]interface{}, rawManifest map[string]interface{}, ) *WorkloadV2`

NewWorkloadV2 instantiates a new WorkloadV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2WithDefaults

`func NewWorkloadV2WithDefaults() *WorkloadV2`

NewWorkloadV2WithDefaults instantiates a new WorkloadV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *WorkloadV2) GetMetadata() WorkloadV2Metadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *WorkloadV2) GetMetadataOk() (*WorkloadV2Metadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *WorkloadV2) SetMetadata(v WorkloadV2Metadata)`

SetMetadata sets Metadata field to given value.


### GetDesiredPhase

`func (o *WorkloadV2) GetDesiredPhase() DesiredPhase`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *WorkloadV2) GetDesiredPhaseOk() (*DesiredPhase, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *WorkloadV2) SetDesiredPhase(v DesiredPhase)`

SetDesiredPhase sets DesiredPhase field to given value.


### GetWorkloadManifest

`func (o *WorkloadV2) GetWorkloadManifest() map[string]interface{}`

GetWorkloadManifest returns the WorkloadManifest field if non-nil, zero value otherwise.

### GetWorkloadManifestOk

`func (o *WorkloadV2) GetWorkloadManifestOk() (*map[string]interface{}, bool)`

GetWorkloadManifestOk returns a tuple with the WorkloadManifest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadManifest

`func (o *WorkloadV2) SetWorkloadManifest(v map[string]interface{})`

SetWorkloadManifest sets WorkloadManifest field to given value.


### GetRawManifest

`func (o *WorkloadV2) GetRawManifest() map[string]interface{}`

GetRawManifest returns the RawManifest field if non-nil, zero value otherwise.

### GetRawManifestOk

`func (o *WorkloadV2) GetRawManifestOk() (*map[string]interface{}, bool)`

GetRawManifestOk returns a tuple with the RawManifest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRawManifest

`func (o *WorkloadV2) SetRawManifest(v map[string]interface{})`

SetRawManifest sets RawManifest field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


