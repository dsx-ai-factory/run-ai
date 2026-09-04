# ClusterControlPlaneConfigMetricsRemoteWrite

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Url** | Pointer to **string** | Remote write url. | [optional] 
**OrgID** | Pointer to **string** | Remote write org ID. | [optional] 
**Credentials** | Pointer to [**ClusterControlPlaneConfigMetricsRemoteWriteCredentials**](ClusterControlPlaneConfigMetricsRemoteWriteCredentials.md) |  | [optional] 

## Methods

### NewClusterControlPlaneConfigMetricsRemoteWrite

`func NewClusterControlPlaneConfigMetricsRemoteWrite() *ClusterControlPlaneConfigMetricsRemoteWrite`

NewClusterControlPlaneConfigMetricsRemoteWrite instantiates a new ClusterControlPlaneConfigMetricsRemoteWrite object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterControlPlaneConfigMetricsRemoteWriteWithDefaults

`func NewClusterControlPlaneConfigMetricsRemoteWriteWithDefaults() *ClusterControlPlaneConfigMetricsRemoteWrite`

NewClusterControlPlaneConfigMetricsRemoteWriteWithDefaults instantiates a new ClusterControlPlaneConfigMetricsRemoteWrite object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUrl

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetOrgID

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) GetOrgID() string`

GetOrgID returns the OrgID field if non-nil, zero value otherwise.

### GetOrgIDOk

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) GetOrgIDOk() (*string, bool)`

GetOrgIDOk returns a tuple with the OrgID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgID

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) SetOrgID(v string)`

SetOrgID sets OrgID field to given value.

### HasOrgID

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) HasOrgID() bool`

HasOrgID returns a boolean if a field has been set.

### GetCredentials

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) GetCredentials() ClusterControlPlaneConfigMetricsRemoteWriteCredentials`

GetCredentials returns the Credentials field if non-nil, zero value otherwise.

### GetCredentialsOk

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) GetCredentialsOk() (*ClusterControlPlaneConfigMetricsRemoteWriteCredentials, bool)`

GetCredentialsOk returns a tuple with the Credentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentials

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) SetCredentials(v ClusterControlPlaneConfigMetricsRemoteWriteCredentials)`

SetCredentials sets Credentials field to given value.

### HasCredentials

`func (o *ClusterControlPlaneConfigMetricsRemoteWrite) HasCredentials() bool`

HasCredentials returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


