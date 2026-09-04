# MappingDeleteSpecSelectorsReq

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadType** | [**GVK**](GVK.md) | The Kubernetes group, version, and kind of the workload type whose mapping the spec selectors are deleted from. | 
**SpecSelectors** | **[]string** | The names of the spec selectors to delete (1-50 names). | 

## Methods

### NewMappingDeleteSpecSelectorsReq

`func NewMappingDeleteSpecSelectorsReq(workloadType GVK, specSelectors []string, ) *MappingDeleteSpecSelectorsReq`

NewMappingDeleteSpecSelectorsReq instantiates a new MappingDeleteSpecSelectorsReq object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMappingDeleteSpecSelectorsReqWithDefaults

`func NewMappingDeleteSpecSelectorsReqWithDefaults() *MappingDeleteSpecSelectorsReq`

NewMappingDeleteSpecSelectorsReqWithDefaults instantiates a new MappingDeleteSpecSelectorsReq object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadType

`func (o *MappingDeleteSpecSelectorsReq) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *MappingDeleteSpecSelectorsReq) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *MappingDeleteSpecSelectorsReq) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.


### GetSpecSelectors

`func (o *MappingDeleteSpecSelectorsReq) GetSpecSelectors() []string`

GetSpecSelectors returns the SpecSelectors field if non-nil, zero value otherwise.

### GetSpecSelectorsOk

`func (o *MappingDeleteSpecSelectorsReq) GetSpecSelectorsOk() (*[]string, bool)`

GetSpecSelectorsOk returns a tuple with the SpecSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelectors

`func (o *MappingDeleteSpecSelectorsReq) SetSpecSelectors(v []string)`

SetSpecSelectors sets SpecSelectors field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


