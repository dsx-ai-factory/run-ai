# RealmReconcileRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Force** | Pointer to **bool** |  | [optional] 
**ReconcilerType** | Pointer to [**NullableReconcilerType**](ReconcilerType.md) |  | [optional] 

## Methods

### NewRealmReconcileRequest

`func NewRealmReconcileRequest() *RealmReconcileRequest`

NewRealmReconcileRequest instantiates a new RealmReconcileRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRealmReconcileRequestWithDefaults

`func NewRealmReconcileRequestWithDefaults() *RealmReconcileRequest`

NewRealmReconcileRequestWithDefaults instantiates a new RealmReconcileRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetForce

`func (o *RealmReconcileRequest) GetForce() bool`

GetForce returns the Force field if non-nil, zero value otherwise.

### GetForceOk

`func (o *RealmReconcileRequest) GetForceOk() (*bool, bool)`

GetForceOk returns a tuple with the Force field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForce

`func (o *RealmReconcileRequest) SetForce(v bool)`

SetForce sets Force field to given value.

### HasForce

`func (o *RealmReconcileRequest) HasForce() bool`

HasForce returns a boolean if a field has been set.

### GetReconcilerType

`func (o *RealmReconcileRequest) GetReconcilerType() ReconcilerType`

GetReconcilerType returns the ReconcilerType field if non-nil, zero value otherwise.

### GetReconcilerTypeOk

`func (o *RealmReconcileRequest) GetReconcilerTypeOk() (*ReconcilerType, bool)`

GetReconcilerTypeOk returns a tuple with the ReconcilerType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReconcilerType

`func (o *RealmReconcileRequest) SetReconcilerType(v ReconcilerType)`

SetReconcilerType sets ReconcilerType field to given value.

### HasReconcilerType

`func (o *RealmReconcileRequest) HasReconcilerType() bool`

HasReconcilerType returns a boolean if a field has been set.

### SetReconcilerTypeNil

`func (o *RealmReconcileRequest) SetReconcilerTypeNil(b bool)`

 SetReconcilerTypeNil sets the value for ReconcilerType to be an explicit nil

### UnsetReconcilerType
`func (o *RealmReconcileRequest) UnsetReconcilerType()`

UnsetReconcilerType ensures that no value is present for ReconcilerType, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


