# WorkloadKubernetesRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Group** | Pointer to **string** | The API group of the Kubernetes resource (for example, \&quot;ray.io\&quot;). | [optional] 
**Version** | Pointer to **string** | The API version of the Kubernetes resource (for example, \&quot;v1\&quot;). | [optional] 

## Methods

### NewWorkloadKubernetesRef

`func NewWorkloadKubernetesRef() *WorkloadKubernetesRef`

NewWorkloadKubernetesRef instantiates a new WorkloadKubernetesRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadKubernetesRefWithDefaults

`func NewWorkloadKubernetesRefWithDefaults() *WorkloadKubernetesRef`

NewWorkloadKubernetesRefWithDefaults instantiates a new WorkloadKubernetesRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGroup

`func (o *WorkloadKubernetesRef) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *WorkloadKubernetesRef) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *WorkloadKubernetesRef) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *WorkloadKubernetesRef) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

### GetVersion

`func (o *WorkloadKubernetesRef) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadKubernetesRef) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadKubernetesRef) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadKubernetesRef) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


