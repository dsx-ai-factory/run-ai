# WorkloadV2UpdateConfigurationPvcMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | PVC name. Must be unique across the workload&#39;s PVCs and conform to Kubernetes DNS-1123 subdomain rules.  | 
**Labels** | Pointer to **map[string]string** | Labels to apply to the created PVC. Map of string keys and string values.  | [optional] 
**Annotations** | Pointer to **map[string]string** | Annotations to apply to the created PVC. Map of string keys and string values. Commonly used to pass CSI driver hints.  | [optional] 

## Methods

### NewWorkloadV2UpdateConfigurationPvcMetadata

`func NewWorkloadV2UpdateConfigurationPvcMetadata(name string, ) *WorkloadV2UpdateConfigurationPvcMetadata`

NewWorkloadV2UpdateConfigurationPvcMetadata instantiates a new WorkloadV2UpdateConfigurationPvcMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2UpdateConfigurationPvcMetadataWithDefaults

`func NewWorkloadV2UpdateConfigurationPvcMetadataWithDefaults() *WorkloadV2UpdateConfigurationPvcMetadata`

NewWorkloadV2UpdateConfigurationPvcMetadataWithDefaults instantiates a new WorkloadV2UpdateConfigurationPvcMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) SetName(v string)`

SetName sets Name field to given value.


### GetLabels

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) GetAnnotations() map[string]string`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) GetAnnotationsOk() (*map[string]string, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) SetAnnotations(v map[string]string)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *WorkloadV2UpdateConfigurationPvcMetadata) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


