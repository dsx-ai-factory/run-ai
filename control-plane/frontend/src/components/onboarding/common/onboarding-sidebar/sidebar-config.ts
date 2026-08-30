import { showsClusterLocationStep } from "@/services/infra/boot-config.service/boot-config.service";
import type { OnboardingSidebarConfig } from "./types";
import { OnboardingStep } from "@/components/onboarding/models/admin-onboarding.model";

// Build the cluster setup flow from the configured capabilities.
const _getClusterSetupSubSteps = () => {
  const steps = [
    { id: 1, title: "Cluster name", isClickable: true },
    { id: 2, title: "Cluster URL", isClickable: true },
    { id: 3, title: "Installation instructions", isClickable: true },
  ];

  if (showsClusterLocationStep()) {
    // Insert the optional location step while preserving sequential IDs.
    steps.splice(1, 0, { id: 2, title: "Cluster location", isClickable: true });
    steps.forEach((step, index) => (step.id = index + 1));
  }

  return steps;
};

export const ONBOARDING_SIDEBAR_CONFIG: OnboardingSidebarConfig = {
  steps: [
    {
      id: OnboardingStep.ClusterSetup,
      title: "Install your cluster",
      subSteps: _getClusterSetupSubSteps(),
      isClickable: true,
    },
    {
      id: OnboardingStep.SsoSetup,
      title: "Set up SSO (Single Sign-On)",
      subSteps: [
        {
          id: 1,
          title: "Protocol",
          isClickable: true,
        },
        {
          id: 2,
          title: "IdP metadata",
          isClickable: true,
        },
        {
          id: 3,
          title: "NVIDIA Run:ai metadata",
          isClickable: true,
        },
      ],
      isClickable: true,
    },
    {
      id: OnboardingStep.EmailServerSetup,
      title: "Set up an email server",
      isClickable: true,
    },
    {
      id: OnboardingStep.TeamSetup,
      title: "Create your first research team",
      subSteps: [
        {
          id: 1,
          title: "Team name & members",
          isClickable: false,
        },
        {
          id: 2,
          title: "Team access",
          isClickable: false,
        },
      ],
      isClickable: true,
    },
  ],
};
