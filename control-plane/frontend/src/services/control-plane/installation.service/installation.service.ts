import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import type { ServiceAccountPostResponse } from "@/swagger-models/identity-manager-client";
import { InstallationErrorType } from "@/components/onboarding/models/admin-onboarding.model";
import type { InstallationResult } from "@/components/onboarding/models/admin-onboarding.model";

// Helper function to determine error type based on error content
const determineErrorType = (error: unknown): InstallationErrorType => {
  const errorMessage = error instanceof Error ? error.message.toLowerCase() : "";

  if (errorMessage.includes("installation") || errorMessage.includes("instructions")) {
    return InstallationErrorType.InstallationInstructions;
  }

  // Default to cluster creation error
  return InstallationErrorType.ClusterCreation;
};

// Function to find existing cluster by name
const findExistingCluster = async (clusterName: string): Promise<DisplayedCluster | null> => {
  try {
    const clusters = await clusterService.list();

    // Find cluster by name (case-insensitive)
    const existingCluster = clusters.find((cluster) => cluster.name?.toLowerCase() === clusterName.toLowerCase());

    return existingCluster || null;
  } catch (err: unknown) {
    console.error("Error finding existing cluster:", err);
    return null;
  }
};

// Function to create or find cluster
const createOrFindCluster = async (
  clusterName: string,
): Promise<{
  cluster: DisplayedCluster;
  isExisting: boolean;
}> => {
  try {
    // First, check if cluster already exists
    const existingCluster = await findExistingCluster(clusterName);

    if (existingCluster) {
      return {
        cluster: existingCluster,
        isExisting: true,
      };
    } else {
      const createdCluster = await clusterService.create(clusterName);

      if (!createdCluster?.uuid) {
        throw new Error("Failed to create cluster");
      }

      return {
        cluster: createdCluster,
        isExisting: false,
      };
    }
  } catch (error) {
    console.error("Error creating or finding cluster:", error);
    throw error;
  }
};

// Function to get client secret
const getClientSecret = async (): Promise<string> => {
  try {
    const clusterControlPlaneConfigAgentClient: ServiceAccountPostResponse =
      await serviceAccountsService.getRunAiInstallerServiceAccountClient();

    return clusterControlPlaneConfigAgentClient?.clientSecret || "";
  } catch (error) {
    console.error("Error getting client secret:", error);
    throw error;
  }
};

// Function to get installation instructions
const getInstallationInstructions = async (
  cluster: DisplayedCluster,
  clusterVersion: string,
  clusterUrl: string,
): Promise<string> => {
  try {
    if (!cluster?.uuid) {
      throw new Error("Cluster UUID is required for installation instructions");
    }

    const response = await clusterService.getClusterInstallationInfo(cluster.uuid, clusterVersion, clusterUrl);

    if (response?.data?.installationStr) {
      return response.data.installationStr;
    }

    throw new Error("Installation instructions not available");
  } catch (error) {
    console.error("Error fetching installation instructions:", error);
    throw error;
  }
};

// Main function to prepare installation
const prepareInstallation = async (
  clusterName: string,
  clusterVersion: string,
  clusterUrl: string,
): Promise<InstallationResult> => {
  try {
    // Step 1: Create or find cluster
    const createdCluster = await createOrFindCluster(clusterName);
    // Step 2: Get client secret
    const clientSecret = await getClientSecret();

    // Step 3: Get installation instructions
    const installationInstructions = await getInstallationInstructions(
      createdCluster.cluster,
      clusterVersion,
      clusterUrl,
    );

    const installationCommand = installationInstructions || "";

    return {
      success: true,
      data: {
        createdCluster: createdCluster.cluster,
        clientSecret,
        installationInstructions,
        isExistingCluster: createdCluster.isExisting,
        installationCommand,
      },
    };
  } catch (error) {
    console.error("Error preparing installation:", error);

    // Determine error type based on the error
    const errorType = determineErrorType(error);

    return {
      success: false,
      error: {
        type: errorType,
        message: error instanceof Error ? error.message : "Unknown error occurred",
      },
    };
  }
};

export const installationService = {
  prepareInstallation,
};
