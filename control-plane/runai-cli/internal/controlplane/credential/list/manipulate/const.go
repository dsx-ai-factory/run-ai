package manipulate

type CredentialColumn string

const (
	ColumnKeyName    CredentialColumn = "name"
	ColumnKeyType    CredentialColumn = "type"
	ColumnKeyScope   CredentialColumn = "scope"
	ColumnKeyK8sName CredentialColumn = "k8sname"
	ColumnKeyStatus  CredentialColumn = "status"
)

const (
	DefaultWidth = 20
)
