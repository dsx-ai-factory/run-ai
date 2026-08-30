interface ChartDrillState {
  meta: unknown | null;
}

interface CompletedChartDrilldown<TState> {
  stack: TState[];
  breadcrumbs: string[];
  announcement: string;
}

interface ChartDrillUpNavigation<TState> {
  targetState: TState;
  stack: TState[];
  breadcrumbs: string[];
  announcement: string;
}

function getNextChartDrilldownBreadcrumbs(currentBreadcrumbs: string[], rootLabel: string): string[] {
  return currentBreadcrumbs.length === 0 ? [rootLabel] : [...currentBreadcrumbs];
}

function completeChartDrilldown<TState>(
  stack: TState[],
  previousState: TState,
  nextBreadcrumbs: string[],
  breadcrumbLabel: string,
  categoryCount: number,
): CompletedChartDrilldown<TState> {
  return {
    stack: [...stack, previousState],
    breadcrumbs: [...nextBreadcrumbs, breadcrumbLabel],
    announcement: `${breadcrumbLabel} details loaded. ${categoryCount} categories shown.`,
  };
}

function getChartDrillUpNavigation<TState extends ChartDrillState>(
  stack: TState[],
  breadcrumbs: string[],
  index: number,
): ChartDrillUpNavigation<TState> | null {
  if (index < 0 || index >= stack.length) return null;

  const targetState = stack.at(index);
  if (!targetState) return null;
  const isTopLevel = targetState.meta === null;

  return {
    targetState,
    stack: stack.slice(0, index),
    breadcrumbs: isTopLevel ? [] : breadcrumbs.slice(0, index + 1),
    announcement: isTopLevel ? "Top chart level loaded." : "Previous drilldown level loaded.",
  };
}

export { completeChartDrilldown, getChartDrillUpNavigation, getNextChartDrilldownBreadcrumbs };
