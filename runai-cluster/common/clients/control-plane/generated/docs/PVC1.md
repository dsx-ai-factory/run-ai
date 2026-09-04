# PVC1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | name of the pvc. | 
**Origin** | Pointer to [**NullableOrigin**](Origin.md) |  | [optional] 
**Issues** | Pointer to [**[]Issue**](Issue.md) | list of issues that was found while syncing the resource to the cluster | [optional] 

## Methods

### NewPVC1

`func NewPVC1(name string, ) *PVC1`

NewPVC1 instantiates a new PVC1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVC1WithDefaults

`func NewPVC1WithDefaults() *PVC1`

NewPVC1WithDefaults instantiates a new PVC1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PVC1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PVC1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PVC1) SetName(v string)`

SetName sets Name field to given value.


### GetOrigin

`func (o *PVC1) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *PVC1) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *PVC1) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.

### HasOrigin

`func (o *PVC1) HasOrigin() bool`

HasOrigin returns a boolean if a field has been set.

### SetOriginNil

`func (o *PVC1) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *PVC1) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetIssues

`func (o *PVC1) GetIssues() []Issue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *PVC1) GetIssuesOk() (*[]Issue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *PVC1) SetIssues(v []Issue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *PVC1) HasIssues() bool`

HasIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


