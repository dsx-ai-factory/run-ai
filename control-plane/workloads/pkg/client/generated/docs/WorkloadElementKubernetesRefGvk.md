# WorkloadElementKubernetesRefGvk

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Group** | Pointer to **string** | The API group of the Kubernetes resource (for example, apps or batch). | [optional] 
**Version** | Pointer to **string** | The API version of the Kubernetes resource (for example, v1 or v1beta1). | [optional] 
**Kind** | Pointer to **string** | The kind of the Kubernetes resource (for example, Deployment or Job). | [optional] 

## Methods

### NewWorkloadElementKubernetesRefGvk

`func NewWorkloadElementKubernetesRefGvk() *WorkloadElementKubernetesRefGvk`

NewWorkloadElementKubernetesRefGvk instantiates a new WorkloadElementKubernetesRefGvk object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadElementKubernetesRefGvkWithDefaults

`func NewWorkloadElementKubernetesRefGvkWithDefaults() *WorkloadElementKubernetesRefGvk`

NewWorkloadElementKubernetesRefGvkWithDefaults instantiates a new WorkloadElementKubernetesRefGvk object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGroup

`func (o *WorkloadElementKubernetesRefGvk) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *WorkloadElementKubernetesRefGvk) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *WorkloadElementKubernetesRefGvk) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *WorkloadElementKubernetesRefGvk) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

### GetVersion

`func (o *WorkloadElementKubernetesRefGvk) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadElementKubernetesRefGvk) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadElementKubernetesRefGvk) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadElementKubernetesRefGvk) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetKind

`func (o *WorkloadElementKubernetesRefGvk) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *WorkloadElementKubernetesRefGvk) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *WorkloadElementKubernetesRefGvk) SetKind(v string)`

SetKind sets Kind field to given value.

### HasKind

`func (o *WorkloadElementKubernetesRefGvk) HasKind() bool`

HasKind returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


