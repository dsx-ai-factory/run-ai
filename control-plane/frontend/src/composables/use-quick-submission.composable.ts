import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";

import {
  WorkloadType,
  type WorkspaceCreationRequest,
  type TrainingCreationRequest,
  type DistributedCreationRequest,
  type InferenceCreationRequest,
  type Training,
  type Inference,
  type Workspace,
} from "@/swagger-models/workloads-client";
import { alertUtil } from "@/utils/alert.util";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { ref } from "vue";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { useWorkloadStore } from "@/stores/workload.store";
import { HttpErrorResponse } from "@/models/http-response.model";

type TQuickSubmissionRequest =
  | WorkspaceCreationRequest
  | TrainingCreationRequest
  | InferenceCreationRequest
  | DistributedCreationRequest;

const workloadCreationByType: Record<
  WorkloadType,
  (req: TQuickSubmissionRequest) => Promise<Workspace | Training | Inference | null>
> = {
  [WorkloadType.Workspace]: async (req) => {
    return await workspaceService.createWorkspaceFlat(req);
  },
  [WorkloadType.Training]: async (req) => {
    return await trainingService.createTrainingFlat(req);
  },
  [WorkloadType.Distributed]: async (req) => {
    return await trainingService.createDistributedFlat(req);
  },
  [WorkloadType.Inference]: async (req) => {
    return await inferenceService.createInferenceFlatApi(req);
  },
  [WorkloadType.DistributedInference]: async (req) => {
    // placeholder: not sure when or if this is supposed to be supported in the UI
    console.log(req);
    throw new Error("Distributed inference is not supported for quick submission");
  },
};

export function useQuickSubmission() {
  const $q = useQuasar();
  const $router = useRouter();
  const isLoading = ref(false);

  const onQuickSubmit = async (type: WorkloadType, req: TQuickSubmissionRequest): Promise<void> => {
    try {
      isLoading.value = true;
      const created = await workloadCreationByType[type](req);

      if (created) useWorkloadStore().setWorkloadAdded(created);
      $router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query: { clusterId: req.clusterId } });
      $q.notify(alertUtil.getSuccess(`${type} submitted successfully`));
    } catch (error: unknown) {
      if (error instanceof HttpErrorResponse) {
        console.error(error.serialize());
        $q.notify(alertUtil.getError(error.message));
      } else {
        console.error(error);
        $q.notify(alertUtil.getError(`Failed to submit ${type.toLocaleLowerCase()}`));
      }
    } finally {
      isLoading.value = false;
    }
  };

  return {
    isLoading,
    onQuickSubmit,
  };
}
