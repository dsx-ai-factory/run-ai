package datavolume

const (
	ProjectsDepartmentFieldIndexName = "spec.department"

	KaiProjectsParentFieldIndexName = "spec.parent"

	FinalizerName = "datavolume.run.ai/finalize"
)

type DatavolumePhaseMessage string

const (
	ProjectOfOriginalPvcNotFound   DatavolumePhaseMessage = "ProjectOfOriginalPvcNotFound"
	NamespaceOfOriginalPvcNotFound DatavolumePhaseMessage = "NamespaceOfOriginalPvcNotFound"
	ScopeNotFound                  DatavolumePhaseMessage = "ScopeNotFound"
	//TODO: perhaps later, remove this "internalError" thing..
	InternalError              DatavolumePhaseMessage = "InternalError"
	OriginalPvcDoesntExist     DatavolumePhaseMessage = "OriginalPvcDoesntExist"
	OriginalPvcNotBound        DatavolumePhaseMessage = "OriginalPvcNotBound"
	PvOfOriginalPvcDoesntExist DatavolumePhaseMessage = "PvOfOriginalPvcDoesntExist"
	PvOfOriginalPvcNotBound    DatavolumePhaseMessage = "PvOfOriginalPvcNotBound"
	FailedToCreatePvcsPvs      DatavolumePhaseMessage = "FailedToCreatePvcsPvs"
	FailedToUpdatePvcsPvs      DatavolumePhaseMessage = "FailedToUpdatePvcsPvs"
	FailedToDeletePvcsPvs      DatavolumePhaseMessage = "FailedToDeletePvcsPvs"
)
