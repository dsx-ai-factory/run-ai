# KubernetesPermissions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PredefinedRole** | Pointer to **NullableString** | The ID of an NVIDIA Run:ai predefined role. The custom role will include the mapped Kubernetes ClusterRole and its permissions. | [optional] 

## Methods

### NewKubernetesPermissions

`func NewKubernetesPermissions() *KubernetesPermissions`

NewKubernetesPermissions instantiates a new KubernetesPermissions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewKubernetesPermissionsWithDefaults

`func NewKubernetesPermissionsWithDefaults() *KubernetesPermissions`

NewKubernetesPermissionsWithDefaults instantiates a new KubernetesPermissions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPredefinedRole

`func (o *KubernetesPermissions) GetPredefinedRole() string`

GetPredefinedRole returns the PredefinedRole field if non-nil, zero value otherwise.

### GetPredefinedRoleOk

`func (o *KubernetesPermissions) GetPredefinedRoleOk() (*string, bool)`

GetPredefinedRoleOk returns a tuple with the PredefinedRole field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPredefinedRole

`func (o *KubernetesPermissions) SetPredefinedRole(v string)`

SetPredefinedRole sets PredefinedRole field to given value.

### HasPredefinedRole

`func (o *KubernetesPermissions) HasPredefinedRole() bool`

HasPredefinedRole returns a boolean if a field has been set.

### SetPredefinedRoleNil

`func (o *KubernetesPermissions) SetPredefinedRoleNil(b bool)`

 SetPredefinedRoleNil sets the value for PredefinedRole to be an explicit nil

### UnsetPredefinedRole
`func (o *KubernetesPermissions) UnsetPredefinedRole()`

UnsetPredefinedRole ensures that no value is present for PredefinedRole, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


