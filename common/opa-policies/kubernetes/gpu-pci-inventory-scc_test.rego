# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package main

gpu_pci_inventory_scc_fixture = {
    "apiVersion": "security.openshift.io/v1",
    "kind": "SecurityContextConstraints",
    "metadata": {"name": "gpu-pci-inventory"},
    "allowHostDirVolumePlugin": true,
    "allowHostNetwork": false,
    "allowHostPID": false,
    "allowHostIPC": false,
    "allowHostPorts": false,
    "allowPrivilegedContainer": false,
    "allowPrivilegeEscalation": false,
    "defaultAllowPrivilegeEscalation": false,
    "readOnlyRootFilesystem": true,
    "allowedCapabilities": [],
    "requiredDropCapabilities": ["ALL"],
    "allowedHostPaths": [
        {"pathPrefix": "/sys", "readOnly": true},
        {"pathPrefix": "/proc/driver/nvidia/gpus", "readOnly": true},
    ],
    "volumes": ["hostPath", "projected", "secret", "configMap", "downwardAPI"],
    "runAsUser": {"type": "MustRunAs", "uid": 65532},
    "seLinuxContext": {"type": "RunAsAny"},
    "seccompProfiles": ["runtime/default"],
}

test_gpu_pci_inventory_restricted_host_access_is_allowed {
    exception[["allowed_scc"]] with input as gpu_pci_inventory_scc_fixture
}

test_gpu_pci_inventory_name_alone_never_grants_an_exception {
    not exception[["allowed_scc"]] with input as {"metadata": {"name": "gpu-pci-inventory"}}
}

test_gpu_pci_inventory_exception_rejects_unsafe_changes {
    changes := [
        {"apiVersion": "other/v1"},
        {"kind": "Other"},
        {"metadata": {"name": "another-scc"}},
        {"allowPrivilegedContainer": true},
        {"allowPrivilegeEscalation": true},
        {"defaultAllowPrivilegeEscalation": true},
        {"allowHostNetwork": true},
        {"allowHostPID": true},
        {"allowHostIPC": true},
        {"allowHostPorts": true},
        {"readOnlyRootFilesystem": false},
        {"allowedCapabilities": ["SYS_ADMIN"]},
        {"defaultAddCapabilities": ["SYS_ADMIN"]},
        {"requiredDropCapabilities": []},
        {"allowedHostPaths": [{"pathPrefix": "/", "readOnly": true}]},
        {"allowedHostPaths": [
            {"pathPrefix": "/sys", "readOnly": false},
            {"pathPrefix": "/proc/driver/nvidia/gpus", "readOnly": true},
        ]},
        {"allowedHostPaths": [
            {"pathPrefix": "/sys", "readOnly": true},
            {"pathPrefix": "/proc/driver/nvidia/gpus", "readOnly": true},
            {"pathPrefix": "/etc", "readOnly": true},
        ]},
        {"runAsUser": {"type": "RunAsAny"}},
        {"runAsUser": {"type": "MustRunAs", "uid": 0}},
        {"seLinuxContext": {"type": "RunAsAny", "seLinuxOptions": {"type": "spc_t"}}},
        {"volumes": ["*"]},
        {"seccompProfiles": ["*"]},
    ]
    accepted := [change |
        change := changes[_]
        modified := object.union(gpu_pci_inventory_scc_fixture, change)
        exception[["allowed_scc"]] with input as modified
    ]
    count(accepted) == 0
}

test_gpu_pci_inventory_exception_requires_explicit_security_fields {
    fields := ["allowHostNetwork", "allowHostPID", "allowHostIPC", "allowHostPorts",
        "allowPrivilegedContainer", "allowPrivilegeEscalation", "defaultAllowPrivilegeEscalation",
        "readOnlyRootFilesystem", "allowedCapabilities", "requiredDropCapabilities",
        "allowedHostPaths", "runAsUser", "seLinuxContext", "volumes", "seccompProfiles"]
    accepted := [field |
        field := fields[_]
        modified := object.remove(gpu_pci_inventory_scc_fixture, [field])
        exception[["allowed_scc"]] with input as modified
    ]
    count(accepted) == 0
}
