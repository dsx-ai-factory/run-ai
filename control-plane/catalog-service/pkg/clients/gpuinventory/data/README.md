# NVIDIA GPU name / PCI snapshot

`nvidia-gpus.json` is generated from NVIDIA driver-assistant's `supported-gpus.json`, distributed in the official CUDA package repository. The source package URL and verified SHA256 are embedded in the JSON. `LICENSE.nvidia-driver-assistant` preserves NVIDIA's MIT attribution.

Names use exactly GPU Feature Discovery's product-label sanitization: remove characters outside `[A-Za-z0-9_. -]`, then replace runs of spaces with `-`. Reference: NVIDIA/k8s-device-plugin `internal/lm/resource.go`, commit `ad5bc6ccbf9608431c0c94927866882d998a6196`. NVIDIA PCI vendor ID is `10de`; subsystem vendor IDs in the source are not PCI vendor IDs.

Keep every distinct device ID for a name. For example `NVIDIA-B200` has `2901:10de` and `2909:10de`; the name cannot distinguish them. Neither capacity nor the availability of an NGC tested profile can establish its identity. No fuzzy aliases, family-name expansion, or MIG/SHARED suffix stripping is performed.

The service embeds the snapshot and its NVIDIA MIT attribution at build time. No scheduled or runtime download occurs. Unknown or ambiguous names cannot produce a direct/tested match; GPU-agnostic profile fallback remains available.

## Refresh

From `control-plane/catalog-service`, run:

```sh
python3 scripts/refresh_gpu_map.py \
  --package-url https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/x86_64/nvidia-driver-assistant_0.62.71.10-1_all.deb \
  --sha256 68d24b4e8e8812ed5e6a1763e4fee0951e2a1463ab9f3bc7bac1b814edcdd7a2
```

For an update, select a newer driver-assistant package from NVIDIA's official repository and obtain its SHA256 from that repository's `Packages` index. Supply both explicitly. The command verifies the checksum before reading the archive and replaces only the generated JSON. Preserve/update the upstream license if NVIDIA changes it. Review the snapshot diff and run `go test ./pkg/clients/gpuinventory ./pkg/services/rightsizing` before committing and releasing catalog-service. Do not manually edit GPU entries.
