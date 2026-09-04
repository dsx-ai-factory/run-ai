# PodsSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Namespace** | **string** | The namespace we are syncing objects in | 
**Pods** | [**[]Pod**](Pod.md) |  | 
**From** | Pointer to **string** | the first pod id to sync. if empty, all pods before the first one sent will be deleted. | [optional] 
**To** | Pointer to **string** | the last pod id to sync. if empty, all pods after the last one sent will be deleted. | [optional] 

## Methods

### NewPodsSync

`func NewPodsSync(namespace string, pods []Pod, ) *PodsSync`

NewPodsSync instantiates a new PodsSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodsSyncWithDefaults

`func NewPodsSyncWithDefaults() *PodsSync`

NewPodsSyncWithDefaults instantiates a new PodsSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNamespace

`func (o *PodsSync) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *PodsSync) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *PodsSync) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetPods

`func (o *PodsSync) GetPods() []Pod`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *PodsSync) GetPodsOk() (*[]Pod, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *PodsSync) SetPods(v []Pod)`

SetPods sets Pods field to given value.


### GetFrom

`func (o *PodsSync) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *PodsSync) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *PodsSync) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *PodsSync) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *PodsSync) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *PodsSync) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *PodsSync) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *PodsSync) HasTo() bool`

HasTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


