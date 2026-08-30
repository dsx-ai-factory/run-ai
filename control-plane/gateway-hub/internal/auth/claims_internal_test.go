package auth

import (
	"encoding/json"
	"testing"
)

func TestRunaiClaims_Audience(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    []string
		wantErr bool
	}{
		{
			name: "aud as single string",
			json: `{"aud": "runai-cli"}`,
			want: []string{"runai-cli"},
		},
		{
			name: "aud as array",
			json: `{"aud": ["runai-cli", "runai-web"]}`,
			want: []string{"runai-cli", "runai-web"},
		},
		{
			name: "aud missing",
			json: `{}`,
			want: nil,
		},
		{
			name:    "aud invalid type",
			json:    `{"aud": 123}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var claims RunaiClaims
			err := json.Unmarshal([]byte(tt.json), &claims)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (audience: %v)", claims.Audience)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(claims.Audience) != len(tt.want) {
				t.Fatalf("audience: expected %v, got %v", tt.want, claims.Audience)
			}
			for i, v := range tt.want {
				if claims.Audience[i] != v {
					t.Errorf("audience[%d]: expected %q, got %q", i, v, claims.Audience[i])
				}
			}
		})
	}
}
