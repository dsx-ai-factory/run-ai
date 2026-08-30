import type { Workload } from "@/swagger-models/workloads-client";
import { Phase } from "@/swagger-models/workloads-client";
import { storageUtil } from "@/utils/storage.util";
import { dateUtil } from "@/utils/date.util";

interface ILocalPhase {
  oldPhase: string | undefined;
  localPhase: Phase;
  creationTimestamp: number;
}

const WORKLOAD_PHASE_STORAGE_KEY = "workload-local-phase";

export const workloadLocalPhaseUpdater = {
  updateLocalPhases,
  updateWorkloadsByLocalPhases,
};
function updateLocalPhases(workload: Workload, localPhase: Phase): void {
  const currentPhase: Record<string, ILocalPhase> | null = _getWorkloadLocalPhase();
  _saveWorkloadPhase({
    ...currentPhase,
    [workload.id]: {
      oldPhase: workload.phase,
      localPhase: localPhase,
      creationTimestamp: Date.now(),
    },
  });
}
function updateWorkloadsByLocalPhases(workloads: Array<Workload>): Array<Workload> {
  const currentPhase: Record<string, ILocalPhase> | null = _getWorkloadLocalPhase();
  if (!currentPhase) return workloads;

  return workloads.map((workload: Workload) => _updateWorkloadPhase(workload));
}

function _updateWorkloadPhase(workload: Workload): Workload {
  const currentPhase: Record<string, ILocalPhase> | null = _getWorkloadLocalPhase();
  if (!currentPhase) return workload;
  const workloadToCheck: ILocalPhase | undefined = currentPhase[workload.id];

  if (!workloadToCheck) return workload;
  if (shouldRemoveLocalPhase(workloadToCheck, workload)) {
    delete currentPhase[workload.id];
    _saveWorkloadPhase(currentPhase);
  } else if (workloadToCheck.localPhase) {
    workload.phase = workloadToCheck.localPhase;
  }
  return workload;
}

function _cleanupPhase() {
  const currentPhase: Record<string, ILocalPhase> | null = _getWorkloadLocalPhase();
  if (!currentPhase || Object.keys(currentPhase).length === 0) return;
  Object.keys(currentPhase).forEach((key: string) => {
    if (_shouldCleanLocalPhase(currentPhase[key])) {
      delete currentPhase[key];
    }
  });
  _saveWorkloadPhase(currentPhase);
}

// this should be called only once in a while,
// because it's a heavy operation and not needed to be called on every phase update
_cleanupPhase();

function _saveWorkloadPhase(localPhase: Record<string, ILocalPhase>): void {
  storageUtil.save<Record<string, ILocalPhase>>(WORKLOAD_PHASE_STORAGE_KEY, localPhase);
}

function shouldRemoveLocalPhase(workloadToCheck: ILocalPhase, workload: Workload): boolean {
  if (workloadToCheck.oldPhase === Phase.Resuming && workload.phase === Phase.Stopped) return false;
  if (workloadToCheck.oldPhase === Phase.Stopping && workload.phase === Phase.Running) return false;
  if (workloadToCheck.oldPhase === Phase.Running && workload.phase === Phase.Stopping) return false;
  if (
    workloadToCheck.oldPhase === Phase.Deleting &&
    (workload.phase === Phase.Running || workload.phase === Phase.Stopped)
  )
    return false;
  return workloadToCheck.oldPhase !== workload.phase || dateUtil.moreThanHourFromNow(workloadToCheck.creationTimestamp);
}

function _shouldCleanLocalPhase(workloadToCheck: ILocalPhase): boolean {
  return dateUtil.moreThanHourFromNow(workloadToCheck.creationTimestamp);
}
function _getWorkloadLocalPhase(): Record<string, ILocalPhase> | null {
  return storageUtil.get<Record<string, ILocalPhase>>(WORKLOAD_PHASE_STORAGE_KEY);
}
