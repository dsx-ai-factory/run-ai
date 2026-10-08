/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

const (
	UserAgentName = "RunAI-Catalog-Service"

	// NGC
	InternalNgcIntegrationName  = "NGCR"
	NgcAuthenticatedTagsBaseURL = "https://api.ngc.nvidia.com/v2/org"
	// For more information about this API, see https://docs.api.nvidia.com/ngc/reference/getguestteamrepoimagelist_1-1
	NgcGuestTagsBaseURL = "https://api.ngc.nvidia.com/v2/repos"
	// For more information about this API, see https://docs.api.nvidia.com/ngc/reference/searchcatalogbyresource
	NgcImagesBaseURL = "https://api.ngc.nvidia.com/v2/search/catalog/resources/CONTAINER"
	// NGC image prefix
	NgcImagePrefix = "nvcr.io/"
)
