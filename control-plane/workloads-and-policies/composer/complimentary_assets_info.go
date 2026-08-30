package composer

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	ComplimentaryAssetInfoAnn = "run.ai/assets-info"

	DefaultSecretDataKeyAccessKeyId     = "AccessKeyId"
	DefaultSecretDataKeySecretAccessKey = "SecretKey"

	AssetKind_HOST_PATH     = "hostPath"
	AssetKind_NFS           = "nfs"
	AssetKind_PVC           = "pvc"
	AssetKind_GIT           = "git"
	AssetKind_S3            = "s3"
	AssetKind_CONFIG_MAP    = "config-map"
	AssetKind_SECRET_VOLUME = "secret-volume"
)

type MapAssetKindToResources map[string][]workloadApi.Datasource

type ComplimentaryAssetInfo struct {
	Environment workloadApi.Environment `json:"environment,omitempty"`

	Datasources MapAssetKindToResources `json:"datasources,omitempty"`
}

func LoadComplimentaryAssetInfo(annotations []workloadApi.Annotation, encoded bool) (*ComplimentaryAssetInfo, *workloadApi.Error) {
	if annotations == nil {
		return nil, nil
	}

	var infoFromAnnotation string

	for _, annotation := range annotations {
		if convutils.Deref(annotation.Name.Get(), "") == ComplimentaryAssetInfoAnn {
			infoFromAnnotation = convutils.Deref(annotation.Value.Get(), "")
			break
		}
	}

	if infoFromAnnotation == "" {
		return nil, nil
	}

	var decodedInfo []byte

	if encoded {
		var decodeErr error

		decodedInfo, decodeErr = b64.StdEncoding.DecodeString(infoFromAnnotation)
		if decodeErr != nil {
			return nil, workloadApi.NewError(http.StatusBadRequest, fmt.Sprintf("failed to base64 decode %s annotation", ComplimentaryAssetInfoAnn))
		}
	} else {
		decodedInfo = []byte(infoFromAnnotation)
	}

	var result ComplimentaryAssetInfo

	err := json.Unmarshal(decodedInfo, &result)
	if err != nil {
		return nil, workloadApi.NewError(http.StatusBadGateway, fmt.Sprintf("failed to parse %s: %s", ComplimentaryAssetInfoAnn, decodedInfo))
	}

	return &result, nil
}
