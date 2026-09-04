#!/bin/sh

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


set -efu

manifest_path="${1:?Usage: download-distribution-assets.sh <manifest-path> <output-directory>}"
output_directory="${2:?Usage: download-distribution-assets.sh <manifest-path> <output-directory>}"

for command_name in curl jq sha256sum; do
    command -v "${command_name}" >/dev/null || {
        printf 'Required command is unavailable: %s\n' "${command_name}" >&2
        exit 1
    }
done

test -f "${manifest_path}" || {
    printf 'Distribution-assets manifest does not exist: %s\n' "${manifest_path}" >&2
    exit 1
}

jq -e '
  .version == 1
  and (.assets | type == "array" and length > 0)
  and all(.assets[];
    (.id | type == "string" and test("^[a-z0-9][a-z0-9._-]*$"))
    and (.url | type == "string" and test("^https://[^\t\r\n]+$"))
    and (.target
      | type == "string"
      and test("^[A-Za-z0-9._/-]+$")
      and (startswith("/") | not)
      and (test("(^|/)\\.\\.(/|$)") | not))
    and (.sha256 | type == "string" and test("^[0-9a-f]{64}$"))
    and (.bytes | type == "number" and floor == . and . > 0)
    and (.contentType
      | type == "string"
      and test("^[A-Za-z0-9.+-]+/[A-Za-z0-9.+-]+$")))
  and ([.assets[].id] | length == (unique | length))
  and ([.assets[].target] | length == (unique | length))
' "${manifest_path}" >/dev/null || {
    printf 'Invalid distribution-assets manifest: %s\n' "${manifest_path}" >&2
    exit 1
}

mkdir -p "${output_directory}"

manifest_rows="$(mktemp "${TMPDIR:-/tmp}/distribution-assets.XXXXXX")"
trap 'rm -f "${manifest_rows}"' 0
trap 'exit 1' HUP INT TERM

jq -r '
  .assets[]
  | [.id, .url, .target, .sha256, (.bytes | tostring), .contentType]
  | @tsv
' "${manifest_path}" >"${manifest_rows}"

tab="$(printf '\t')"
while IFS="${tab}" read -r asset_id source_url target_path expected_sha256 expected_bytes expected_content_type; do
    target_file="${output_directory}/${target_path}"
    temporary_file="${target_file}.download"

    test ! -e "${target_file}" || {
        printf 'Refusing to overwrite distribution asset: %s\n' "${target_file}" >&2
        exit 1
    }

    mkdir -p "$(dirname "${target_file}")"
    rm -f "${temporary_file}"

    printf 'Downloading distribution asset: %s\n' "${asset_id}"
    printf 'Requested URL: %s\n' "${source_url}"

    download_metadata="$(curl \
      --fail --location --show-error --silent \
      --proto '=https' --proto-redir '=https' \
      --retry 3 --retry-all-errors \
      --connect-timeout 10 --max-time 60 \
      --output "${temporary_file}" \
      --write-out '%{url_effective}\t%{http_code}\t%{content_type}\t%{size_download}' \
      "${source_url}")"

    IFS="${tab}" read -r effective_url http_status actual_content_type downloaded_bytes <<EOF
${download_metadata}
EOF

    printf 'Effective URL: %s\n' "${effective_url}"
    printf 'HTTP status: %s\n' "${http_status}"
    printf 'Content type: %s\n' "${actual_content_type}"
    printf 'Downloaded bytes: %s\n' "${downloaded_bytes}"

    test "${http_status}" = "200" || {
        printf 'Unexpected HTTP status for %s: %s\n' "${asset_id}" "${http_status}" >&2
        exit 1
    }

    case "${actual_content_type}" in
        "${expected_content_type}" | "${expected_content_type};"*) ;;
        *)
            printf 'Unexpected content type for %s: %s\n' "${asset_id}" "${actual_content_type}" >&2
            exit 1
            ;;
    esac

    actual_bytes="$(wc -c <"${temporary_file}" | tr -d '[:space:]')"
    test "${actual_bytes}" = "${expected_bytes}" || {
        printf 'Unexpected byte count for %s: expected %s, received %s\n' \
          "${asset_id}" "${expected_bytes}" "${actual_bytes}" >&2
        exit 1
    }

    test "${downloaded_bytes}" = "${expected_bytes}" || {
        printf 'curl byte count does not match the manifest for %s: expected %s, received %s\n' \
          "${asset_id}" "${expected_bytes}" "${downloaded_bytes}" >&2
        exit 1
    }

    actual_sha256="$(sha256sum "${temporary_file}")"
    actual_sha256="${actual_sha256%% *}"
    test "${actual_sha256}" = "${expected_sha256}" || {
        printf 'Unexpected SHA-256 for %s: expected %s, received %s\n' \
          "${asset_id}" "${expected_sha256}" "${actual_sha256}" >&2
        exit 1
    }

    printf 'SHA-256: OK\n'
    mv "${temporary_file}" "${target_file}"
done <"${manifest_rows}"
