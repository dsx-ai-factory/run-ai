/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package display turns the installer's internal machine names (the stable
// identifiers used in JSON events, tests, and control flow) into human-readable
// labels for the text UI. It is the default source of a label; a step, flow, or
// pipeline may override it with a custom title set at construction. JSON/NDJSON
// output keeps the raw machine names untouched.
package display

import "strings"

// acronyms are tokens that must not be sentence-cased. The key is the lower-case
// token as it appears in a kebab machine name; the value is how it is rendered.
var acronyms = map[string]string{
	"cpu":   "CPU",
	"gpu":   "GPU",
	"ram":   "RAM",
	"rbac":  "RBAC",
	"tls":   "TLS",
	"ca":    "CA",
	"nfd":   "NFD",
	"nic":   "NIC",
	"hpa":   "HPA",
	"mpi":   "MPI",
	"lws":   "LWS",
	"nim":   "NIM",
	"sriov": "SR-IOV",
	"ocp":   "OpenShift",
	"api":   "API",
	"dns":   "DNS",
	"url":   "URL",
}

// Humanize is the default label for a kebab machine name. It renders the name
// in sentence case (first word capitalized, the rest lower-cased) while keeping
// known acronyms upper-cased, e.g. "install-gpu-operator" -> "Install GPU
// operator" and "preflight-rbac" -> "Preflight RBAC".
func Humanize(name string) string {
	if name == "" {
		return ""
	}
	words := strings.Split(name, "-")
	for i, w := range words {
		if w == "" {
			continue
		}
		if a, ok := acronyms[strings.ToLower(w)]; ok {
			words[i] = a
			continue
		}
		if i == 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		} else {
			words[i] = strings.ToLower(w)
		}
	}
	return strings.Join(words, " ")
}

// Or returns the custom title when one is set, otherwise the humanized name. It
// backs the struct-based Flow/Pipeline titles (BaseStep resolves its own title
// the same way through this function).
func Or(custom, name string) string {
	if custom != "" {
		return custom
	}
	return Humanize(name)
}
