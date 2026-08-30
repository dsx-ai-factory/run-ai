export interface ITab {
  name: string;
  label: string;
  disabled?: boolean;
  hidden?: boolean;
  errorIcon?: boolean;
  numberingBadge?: string;
  aid: string;
  tabClass?: string | Record<string, boolean>;
}
