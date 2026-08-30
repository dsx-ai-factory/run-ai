package v1alpha1

import (
	"strings"
	"testing"
)

func TestGenerateKartaNameByGvk(t *testing.T) {
	cases := []struct {
		name             string
		group, ver, kind string
		want             string
		wantErr          bool
	}{
		{
			name:  "core gvk (empty group)",
			group: "", ver: "v1", kind: "Pod",
			want: "krt-v1--pod",
		},
		{
			name:  "named group",
			group: "ray.io", ver: "v1", kind: "RayCluster",
			want: "krt-v1-ray-io-raycluster",
		},
		{
			name:  "kubeflow training-operator",
			group: "kubeflow.org", ver: "v1", kind: "TFJob",
			want: "krt-v1-kubeflow-org-tfjob",
		},
		{
			name:  "missing kind",
			group: "x", ver: "v1", kind: "",
			wantErr: true,
		},
		{
			name:  "missing version",
			group: "x", ver: "", kind: "K",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GenerateKartaNameByGvk(tc.group, tc.ver, tc.kind)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestConvertToK8sResourceName_TruncatesAt62(t *testing.T) {
	long := strings.Repeat("a", 80)
	got := convertToK8sResourceName(long)
	if len(got) > kartaNameMaxLen {
		t.Fatalf("expected len <= %d, got %d (%q)", kartaNameMaxLen, len(got), got)
	}
}

func TestConvertToK8sResourceName_LowercasesAndReplacesNonAlnum(t *testing.T) {
	got := convertToK8sResourceName("Foo.BAR_baz")
	want := "foo-bar-baz"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestConvertToK8sResourceName_TrailingDashGetsSuffix(t *testing.T) {
	got := convertToK8sResourceName("foo-")
	want := "foo-a"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
