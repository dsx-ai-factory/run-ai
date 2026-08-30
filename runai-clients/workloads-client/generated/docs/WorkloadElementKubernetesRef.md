# WorkloadElementKubernetesRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **string** | The unique identifier (UID) of the Kubernetes object backing this element. | [optional] 
**Name** | Pointer to **string** | The name of the Kubernetes object backing this element. | [optional] 
**Gvk** | Pointer to [**WorkloadElementKubernetesRefGvk**](WorkloadElementKubernetesRefGvk.md) |  | [optional] 

## Methods

### NewWorkloadElementKubernetesRef

`func NewWorkloadElementKubernetesRef() *WorkloadElementKubernetesRef`

NewWorkloadElementKubernetesRef instantiates a new WorkloadElementKubernetesRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadElementKubernetesRefWithDefaults

`func NewWorkloadElementKubernetesRefWithDefaults() *WorkloadElementKubernetesRef`

NewWorkloadElementKubernetesRefWithDefaults instantiates a new WorkloadElementKubernetesRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *WorkloadElementKubernetesRef) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *WorkloadElementKubernetesRef) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *WorkloadElementKubernetesRef) SetUid(v string)`

SetUid sets Uid field to given value.

### HasUid

`func (o *WorkloadElementKubernetesRef) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetName

`func (o *WorkloadElementKubernetesRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadElementKubernetesRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadElementKubernetesRef) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkloadElementKubernetesRef) HasName() bool`

HasName returns a boolean if a field has been set.

### GetGvk

`func (o *WorkloadElementKubernetesRef) GetGvk() WorkloadElementKubernetesRefGvk`

GetGvk returns the Gvk field if non-nil, zero value otherwise.

### GetGvkOk

`func (o *WorkloadElementKubernetesRef) GetGvkOk() (*WorkloadElementKubernetesRefGvk, bool)`

GetGvkOk returns a tuple with the Gvk field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGvk

`func (o *WorkloadElementKubernetesRef) SetGvk(v WorkloadElementKubernetesRefGvk)`

SetGvk sets Gvk field to given value.

### HasGvk

`func (o *WorkloadElementKubernetesRef) HasGvk() bool`

HasGvk returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


