export interface IRunaiWizardLayoutStep {
  name: string;
  title: string; // Title to show in the stepper
  editTitle?: string; // Display in the edit dropdown in the index page
  done: boolean;
  disabled?: boolean;
  hasError?: boolean;
  tooltip?: string;
  icon?: string; // Custom icon to display instead of step number (e.g., "ph-fill ph-users-three")
}
