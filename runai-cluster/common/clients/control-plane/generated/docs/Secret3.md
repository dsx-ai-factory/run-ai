# Secret3

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | name of the secret. | 
**Type** | [**SecretType**](SecretType.md) |  | 
**Origin** | Pointer to [**NullableOrigin**](Origin.md) |  | [optional] 
**Issues** | Pointer to [**[]Issue**](Issue.md) | list of issues that was found while syncing the resource to the cluster | [optional] 

## Methods

### NewSecret3

`func NewSecret3(name string, type_ SecretType, ) *Secret3`

NewSecret3 instantiates a new Secret3 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecret3WithDefaults

`func NewSecret3WithDefaults() *Secret3`

NewSecret3WithDefaults instantiates a new Secret3 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Secret3) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Secret3) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Secret3) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *Secret3) GetType() SecretType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Secret3) GetTypeOk() (*SecretType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Secret3) SetType(v SecretType)`

SetType sets Type field to given value.


### GetOrigin

`func (o *Secret3) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *Secret3) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *Secret3) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.

### HasOrigin

`func (o *Secret3) HasOrigin() bool`

HasOrigin returns a boolean if a field has been set.

### SetOriginNil

`func (o *Secret3) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *Secret3) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetIssues

`func (o *Secret3) GetIssues() []Issue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *Secret3) GetIssuesOk() (*[]Issue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *Secret3) SetIssues(v []Issue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *Secret3) HasIssues() bool`

HasIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


