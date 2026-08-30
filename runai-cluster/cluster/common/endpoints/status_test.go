package endpoints

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestStatusRawExtensionRoundTrip(t *testing.T) {
	raw, err := NetworkingStatusToRawExtension(&NetworkingStatus{
		ObservedGeneration: 7,
		Endpoints: []EndpointStatus{{
			Phase: PhaseReady,
			Services: []ServiceStatus{{
				Name: "jupyter",
				URLs: []ServiceURLStatus{{
					URL:        "https://jupyter.example.com/",
					SourceKind: ExternalAccessKindIngress,
					SourceName: "jupyter-ing",
				}},
			}},
		}},
	})
	if err != nil {
		t.Fatalf("NetworkingStatusToRawExtension: %v", err)
	}

	status, err := NetworkingStatusFromRawExtension(raw)
	if err != nil {
		t.Fatalf("NetworkingStatusFromRawExtension: %v", err)
	}
	if status.SchemaVersion != StatusSchemaVersionV1 {
		t.Fatalf("schemaVersion: got %q, want %q", status.SchemaVersion, StatusSchemaVersionV1)
	}
	if status.ObservedGeneration != 7 {
		t.Fatalf("observedGeneration: got %d, want 7", status.ObservedGeneration)
	}
	if got := status.Endpoints[0].Services[0].URLs[0].URL; got != "https://jupyter.example.com/" {
		t.Fatalf("url: got %q", got)
	}
}

func TestNetworkingStatusFromRawExtensionRejectsBadSchemaVersion(t *testing.T) {
	_, err := NetworkingStatusFromRawExtension(&runtime.RawExtension{Raw: []byte(`{"schemaVersion":"v9"}`)})
	if !errors.Is(err, ErrUnsupportedSchemaVersion) {
		t.Fatalf("expected ErrUnsupportedSchemaVersion, got %v", err)
	}
}
