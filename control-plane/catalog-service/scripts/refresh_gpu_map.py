#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Generate the bundled lookup from a checksum-verified NVIDIA driver-assistant package."""
import argparse
import hashlib
import io
import json
from pathlib import Path
import re
import tarfile
import urllib.parse
import urllib.request


def generate(package, url, digest):
    if hashlib.sha256(package).hexdigest() != digest.lower():
        raise ValueError("NVIDIA package SHA256 mismatch")
    if not package.startswith(b"!<arch>\n"):
        raise ValueError("Expected a Debian ar archive")
    offset = 8
    while offset + 60 <= len(package):
        header = package[offset:offset + 60]
        if header[58:] != b"`\n":
            raise ValueError("Malformed ar member")
        name = header[:16].decode("ascii").strip().rstrip("/")
        size = int(header[48:58])
        offset += 60
        payload = package[offset:offset + size]
        offset += size + size % 2
        if not name.startswith("data.tar"):
            continue
        with tarfile.open(fileobj=io.BytesIO(payload), mode="r:*") as archive:
            members = [m for m in archive.getmembers() if m.name.lstrip("./") == "usr/share/nvidia-driver-assistant/supported-gpus/supported-gpus.json"]
            if len(members) != 1:
                raise ValueError("Expected exactly one supported-gpus.json")
            source = archive.extractfile(members[0])
            if source is None:
                raise ValueError("GPU database is not a regular file")
            chips = json.load(source)["chips"]
        devices = {}
        for chip in chips:
            # Same transformation as NVIDIA k8s-device-plugin GFD resource.go.
            label = re.sub(" +", "-", re.sub(r"[^A-Za-z0-9_. -]", "", chip["name"]))
            device_id = int(chip["devid"], 16)
            if not label or not 0 <= device_id <= 0xffff:
                raise ValueError("Invalid NVIDIA device entry")
            devices.setdefault(label, set()).add(f"{device_id:04x}:10de")
        if not devices:
            raise ValueError("Empty NVIDIA GPU database")
        return {"source": {"url": url, "sha256": digest.lower()},
                "devices": {name: sorted(ids) for name, ids in sorted(devices.items())}}
    raise ValueError("Missing data archive")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--package-url", required=True)
    parser.add_argument("--sha256", required=True)
    parser.add_argument("--output", type=Path, default=Path(__file__).resolve().parents[1] / "pkg/clients/gpuinventory/data/nvidia-gpus.json")
    args = parser.parse_args()
    url = urllib.parse.urlparse(args.package_url)
    if url.scheme != "https" or url.hostname != "developer.download.nvidia.com" or not re.fullmatch(r"[0-9a-fA-F]{64}", args.sha256):
        parser.error("Use an official NVIDIA HTTPS package URL and SHA256")
    with urllib.request.urlopen(args.package_url, timeout=30) as response:
        package = response.read(16 * 1024 * 1024 + 1)
    if len(package) > 16 * 1024 * 1024:
        raise ValueError("NVIDIA package exceeds size limit")
    snapshot = generate(package, args.package_url, args.sha256)
    args.output.write_text(json.dumps(snapshot, indent=2) + "\n", encoding="utf-8")
    print(f"Generated {len(snapshot['devices'])} product labels in {args.output}")


if __name__ == "__main__":
    main()
