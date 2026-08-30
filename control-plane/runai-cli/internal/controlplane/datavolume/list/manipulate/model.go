package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

type BriefDatavolume struct {
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	OriginPvc string `json:"originPvc"`
	Status    string `json:"status"`
}

type DatavolumesAlignedRow struct {
	Name      string `json:"Name,omitempty" yaml:"Name,omitempty"`
	Scope     string `json:"Scope,omitempty" yaml:"Scope,omitempty"`
	OriginPvc string `json:"Origin Pvc,omitempty" yaml:"Origin Pvc,omitempty"`
	Status    string `json:"Status,omitempty" yaml:"Status,omitempty"`
}

type DatavolumesAlignedData struct {
	Datavolumes []DatavolumesAlignedRow `json:"datavolumes" yaml:"datavolumes"`
}

type Data struct {
	Datavolumes []openapi.DatavolumeNoSharedScopes `json:"datavolumes" yaml:"datavolumes"`
}

type BriefData struct {
	Datavolumes []BriefDatavolume `json:"datavolumes" yaml:"datavolumes"`
}
