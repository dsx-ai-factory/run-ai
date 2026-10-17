// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package chart_inspection reads Helm chart configuration metadata without
// installing charts or evaluating templates.
//
// The service authorizes the caller and resolves saved credentials before calling
// Inspector. Inspector bounds acquisition and parsing, FetchRemote downloads HTTP(S)
// sources, and ParseArchive reads root chart metadata and questions.yaml/questions.yml.
// Uploaded charts use a service-provided fetcher inside the same acquisition limit.
//
// Missing questions return an absent result. Invalid declarations return an
// unusable result with warnings and no partial form; acquisition and archive
// failures return errors. Chart defaults support advisory path checks only.
//
// Only exact-version HTTPS NGC repository selections reuse parsed metadata.
// URLs and uploads are inspected fresh. The cache stores no archive bytes and
// never replaces caller or credential authorization on subsequent requests.
package chart_inspection
