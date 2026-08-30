package main

import "testing"

func TestExpandOutputPathUsesResolvedVersion(t *testing.T) {
	path := expandOutputPath(
		"dist/image-locks/runai-{version}-linux-amd64.yaml",
		"0.0.0-master-10001786",
	)
	expected := "dist/image-locks/runai-0.0.0-master-10001786-linux-amd64.yaml"
	if path != expected {
		t.Fatalf("path = %s", path)
	}
}

func TestExpandOutputPathKeepsLiteralPath(t *testing.T) {
	path := expandOutputPath("custom.yaml", "3.0.0")
	if path != "custom.yaml" {
		t.Fatalf("path = %s", path)
	}
}
