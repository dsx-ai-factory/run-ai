# AIApplicationSyncItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** | The name of the application. The name must be unique within the organization and can only contain lowercase alphanumeric characters and hyphens. It must start with a letter. | [optional] 
**Source** | Pointer to **string** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**Phase** | Pointer to **string** |  | [optional] 
**ProjectId** | Pointer to **string** | The id of the project. | [optional] 
**ProjectName** | Pointer to **string** | The name of the project | [optional] 
**Spec** | Pointer to [**AIApplicationSpec**](AIApplicationSpec.md) |  | [optional] 
**Auth** | Pointer to [**AIApplicationAuth**](AIApplicationAuth.md) |  | [optional] 
**DesiredPhase** | Pointer to **string** |  | [optional] 
**RequiresControlPlaneAuth** | Pointer to **bool** | Set by the control plane for charts it hosts itself (uploaded via chartId, which it rewrites to a control-plane download URL). When true, the cluster application controller attaches its bearer token when downloading the chart. Never set for user-provided chartURL charts.  | [optional] 

## Methods

### NewAIApplicationSyncItem

`func NewAIApplicationSyncItem() *AIApplicationSyncItem`

NewAIApplicationSyncItem instantiates a new AIApplicationSyncItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationSyncItemWithDefaults

`func NewAIApplicationSyncItemWithDefaults() *AIApplicationSyncItem`

NewAIApplicationSyncItemWithDefaults instantiates a new AIApplicationSyncItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AIApplicationSyncItem) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AIApplicationSyncItem) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AIApplicationSyncItem) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *AIApplicationSyncItem) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *AIApplicationSyncItem) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AIApplicationSyncItem) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AIApplicationSyncItem) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *AIApplicationSyncItem) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSource

`func (o *AIApplicationSyncItem) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *AIApplicationSyncItem) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *AIApplicationSyncItem) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *AIApplicationSyncItem) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetType

`func (o *AIApplicationSyncItem) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AIApplicationSyncItem) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AIApplicationSyncItem) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *AIApplicationSyncItem) HasType() bool`

HasType returns a boolean if a field has been set.

### GetPhase

`func (o *AIApplicationSyncItem) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *AIApplicationSyncItem) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *AIApplicationSyncItem) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *AIApplicationSyncItem) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetProjectId

`func (o *AIApplicationSyncItem) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AIApplicationSyncItem) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AIApplicationSyncItem) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *AIApplicationSyncItem) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetProjectName

`func (o *AIApplicationSyncItem) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *AIApplicationSyncItem) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *AIApplicationSyncItem) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *AIApplicationSyncItem) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetSpec

`func (o *AIApplicationSyncItem) GetSpec() AIApplicationSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *AIApplicationSyncItem) GetSpecOk() (*AIApplicationSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *AIApplicationSyncItem) SetSpec(v AIApplicationSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *AIApplicationSyncItem) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetAuth

`func (o *AIApplicationSyncItem) GetAuth() AIApplicationAuth`

GetAuth returns the Auth field if non-nil, zero value otherwise.

### GetAuthOk

`func (o *AIApplicationSyncItem) GetAuthOk() (*AIApplicationAuth, bool)`

GetAuthOk returns a tuple with the Auth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuth

`func (o *AIApplicationSyncItem) SetAuth(v AIApplicationAuth)`

SetAuth sets Auth field to given value.

### HasAuth

`func (o *AIApplicationSyncItem) HasAuth() bool`

HasAuth returns a boolean if a field has been set.

### GetDesiredPhase

`func (o *AIApplicationSyncItem) GetDesiredPhase() string`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *AIApplicationSyncItem) GetDesiredPhaseOk() (*string, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *AIApplicationSyncItem) SetDesiredPhase(v string)`

SetDesiredPhase sets DesiredPhase field to given value.

### HasDesiredPhase

`func (o *AIApplicationSyncItem) HasDesiredPhase() bool`

HasDesiredPhase returns a boolean if a field has been set.

### GetRequiresControlPlaneAuth

`func (o *AIApplicationSyncItem) GetRequiresControlPlaneAuth() bool`

GetRequiresControlPlaneAuth returns the RequiresControlPlaneAuth field if non-nil, zero value otherwise.

### GetRequiresControlPlaneAuthOk

`func (o *AIApplicationSyncItem) GetRequiresControlPlaneAuthOk() (*bool, bool)`

GetRequiresControlPlaneAuthOk returns a tuple with the RequiresControlPlaneAuth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequiresControlPlaneAuth

`func (o *AIApplicationSyncItem) SetRequiresControlPlaneAuth(v bool)`

SetRequiresControlPlaneAuth sets RequiresControlPlaneAuth field to given value.

### HasRequiresControlPlaneAuth

`func (o *AIApplicationSyncItem) HasRequiresControlPlaneAuth() bool`

HasRequiresControlPlaneAuth returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


