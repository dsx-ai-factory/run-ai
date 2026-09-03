# WorkloadV2UpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**WorkloadV2MetadataUpdateParams**](WorkloadV2MetadataUpdateParams.md) |  | [optional] 
**Manifest** | Pointer to **map[string]interface{}** | The full Kubernetes manifest for the workload in JSON form. Mutually exclusive with manifestEncodedBase64. | [optional] 
**ManifestBase64Encoded** | Pointer to **string** | Kubernetes YAML manifest encoded in base 64. Mutually exclusive with manifest. | [optional] 

## Methods

### NewWorkloadV2UpdateRequest

`func NewWorkloadV2UpdateRequest() *WorkloadV2UpdateRequest`

NewWorkloadV2UpdateRequest instantiates a new WorkloadV2UpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2UpdateRequestWithDefaults

`func NewWorkloadV2UpdateRequestWithDefaults() *WorkloadV2UpdateRequest`

NewWorkloadV2UpdateRequestWithDefaults instantiates a new WorkloadV2UpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *WorkloadV2UpdateRequest) GetMetadata() WorkloadV2MetadataUpdateParams`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *WorkloadV2UpdateRequest) GetMetadataOk() (*WorkloadV2MetadataUpdateParams, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *WorkloadV2UpdateRequest) SetMetadata(v WorkloadV2MetadataUpdateParams)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *WorkloadV2UpdateRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetManifest

`func (o *WorkloadV2UpdateRequest) GetManifest() map[string]interface{}`

GetManifest returns the Manifest field if non-nil, zero value otherwise.

### GetManifestOk

`func (o *WorkloadV2UpdateRequest) GetManifestOk() (*map[string]interface{}, bool)`

GetManifestOk returns a tuple with the Manifest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetManifest

`func (o *WorkloadV2UpdateRequest) SetManifest(v map[string]interface{})`

SetManifest sets Manifest field to given value.

### HasManifest

`func (o *WorkloadV2UpdateRequest) HasManifest() bool`

HasManifest returns a boolean if a field has been set.

### GetManifestBase64Encoded

`func (o *WorkloadV2UpdateRequest) GetManifestBase64Encoded() string`

GetManifestBase64Encoded returns the ManifestBase64Encoded field if non-nil, zero value otherwise.

### GetManifestBase64EncodedOk

`func (o *WorkloadV2UpdateRequest) GetManifestBase64EncodedOk() (*string, bool)`

GetManifestBase64EncodedOk returns a tuple with the ManifestBase64Encoded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetManifestBase64Encoded

`func (o *WorkloadV2UpdateRequest) SetManifestBase64Encoded(v string)`

SetManifestBase64Encoded sets ManifestBase64Encoded field to given value.

### HasManifestBase64Encoded

`func (o *WorkloadV2UpdateRequest) HasManifestBase64Encoded() bool`

HasManifestBase64Encoded returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


