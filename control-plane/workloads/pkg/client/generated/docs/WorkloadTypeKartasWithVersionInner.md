# WorkloadTypeKartasWithVersionInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | Pointer to **string** | The version of the resource definition k8s crd. | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The timestamp for the last time the karta was updated. | [optional] 
**Karta** | Pointer to **map[string]interface{}** | The karta for the workload type. | [optional] 

## Methods

### NewWorkloadTypeKartasWithVersionInner

`func NewWorkloadTypeKartasWithVersionInner() *WorkloadTypeKartasWithVersionInner`

NewWorkloadTypeKartasWithVersionInner instantiates a new WorkloadTypeKartasWithVersionInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeKartasWithVersionInnerWithDefaults

`func NewWorkloadTypeKartasWithVersionInnerWithDefaults() *WorkloadTypeKartasWithVersionInner`

NewWorkloadTypeKartasWithVersionInnerWithDefaults instantiates a new WorkloadTypeKartasWithVersionInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVersion

`func (o *WorkloadTypeKartasWithVersionInner) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadTypeKartasWithVersionInner) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadTypeKartasWithVersionInner) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadTypeKartasWithVersionInner) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkloadTypeKartasWithVersionInner) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkloadTypeKartasWithVersionInner) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkloadTypeKartasWithVersionInner) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkloadTypeKartasWithVersionInner) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetKarta

`func (o *WorkloadTypeKartasWithVersionInner) GetKarta() map[string]interface{}`

GetKarta returns the Karta field if non-nil, zero value otherwise.

### GetKartaOk

`func (o *WorkloadTypeKartasWithVersionInner) GetKartaOk() (*map[string]interface{}, bool)`

GetKartaOk returns a tuple with the Karta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKarta

`func (o *WorkloadTypeKartasWithVersionInner) SetKarta(v map[string]interface{})`

SetKarta sets Karta field to given value.

### HasKarta

`func (o *WorkloadTypeKartasWithVersionInner) HasKarta() bool`

HasKarta returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


