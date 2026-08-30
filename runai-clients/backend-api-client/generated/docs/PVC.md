# PVC

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PvcName** | **string** | Identifier of the pvc name. | 

## Methods

### NewPVC

`func NewPVC(pvcName string, ) *PVC`

NewPVC instantiates a new PVC object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCWithDefaults

`func NewPVCWithDefaults() *PVC`

NewPVCWithDefaults instantiates a new PVC object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPvcName

`func (o *PVC) GetPvcName() string`

GetPvcName returns the PvcName field if non-nil, zero value otherwise.

### GetPvcNameOk

`func (o *PVC) GetPvcNameOk() (*string, bool)`

GetPvcNameOk returns a tuple with the PvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvcName

`func (o *PVC) SetPvcName(v string)`

SetPvcName sets PvcName field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


