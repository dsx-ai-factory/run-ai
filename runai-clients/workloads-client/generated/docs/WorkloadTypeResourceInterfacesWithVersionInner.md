# WorkloadTypeResourceInterfacesWithVersionInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | Pointer to **string** | The version of the resource definition k8s crd. | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The timestamp for the last time the resource interface was updated. | [optional] 
**ResourceInterface** | Pointer to **map[string]interface{}** | The resource interface for the workload type. | [optional] 

## Methods

### NewWorkloadTypeResourceInterfacesWithVersionInner

`func NewWorkloadTypeResourceInterfacesWithVersionInner() *WorkloadTypeResourceInterfacesWithVersionInner`

NewWorkloadTypeResourceInterfacesWithVersionInner instantiates a new WorkloadTypeResourceInterfacesWithVersionInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeResourceInterfacesWithVersionInnerWithDefaults

`func NewWorkloadTypeResourceInterfacesWithVersionInnerWithDefaults() *WorkloadTypeResourceInterfacesWithVersionInner`

NewWorkloadTypeResourceInterfacesWithVersionInnerWithDefaults instantiates a new WorkloadTypeResourceInterfacesWithVersionInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVersion

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetResourceInterface

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) GetResourceInterface() map[string]interface{}`

GetResourceInterface returns the ResourceInterface field if non-nil, zero value otherwise.

### GetResourceInterfaceOk

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) GetResourceInterfaceOk() (*map[string]interface{}, bool)`

GetResourceInterfaceOk returns a tuple with the ResourceInterface field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceInterface

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) SetResourceInterface(v map[string]interface{})`

SetResourceInterface sets ResourceInterface field to given value.

### HasResourceInterface

`func (o *WorkloadTypeResourceInterfacesWithVersionInner) HasResourceInterface() bool`

HasResourceInterface returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


