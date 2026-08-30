export interface OnboardingSubStep {
  id: number;
  title: string;
  isClickable?: boolean;
}

export interface OnboardingSidebarStep {
  id: number;
  title: string;
  subSteps?: OnboardingSubStep[];
  isClickable?: boolean;
}

export interface OnboardingSidebarConfig {
  steps: OnboardingSidebarStep[];
}
