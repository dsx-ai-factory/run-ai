/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

const (
	UserAgentName = "RunAI-Catalog-Service"

	// NGC
	InternalNgcIntegrationName = "NGCR"
	// For more information about this API, see https://docs.api.nvidia.com/ngc/reference/getguestteamrepoimagelist_1-1
	NgcAuthenticatedTagsBaseURL = "https://catalog.ngc.nvidia.com/api/containers/images"
	NgcGuestTagsBaseURL         = "https://api.ngc.nvidia.com/v2/repos"
	// For more information about this API, see https://docs.api.nvidia.com/ngc/reference/searchcatalogbyresource
	NgcImagesBaseURL = "https://api.ngc.nvidia.com/v2/search/catalog/resources/CONTAINER"
	// NGC image prefix
	NgcImagePrefix = "nvcr.io/"
)
