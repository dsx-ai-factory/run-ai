import { ServingPortProtocol } from "@/swagger-models/assets-service-client";

// This is an hard-coded uuid in the db migrations therefore we can filter the environments by this id.
export const GPT2_ASSET_ID = "dbf64cad-5652-4698-b24c-b7193230acbe";

export const GPT2_MODEL_ENVIRONMENT_VARIABLE = {
  name: "PATH_PREFIX",
  value: "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
  description:
    "If the cluster uses path-based URLs (e.g., http://my.domain.com/proj/workload), leave the value as is.\n\nIf the cluster uses host-based URLs (e.g., http://proj-workload.my.domain.com), remove this environment variable entirely.",
};

export const GPT2_MODEL_EXPOSED_URL = {
  name: "Chatbot UI",
  container: 3000,
  toolType: "chatbot-ui",
};

export const GPT2_MODEL_SERVING_PORT = {
  protocol: ServingPortProtocol.Http,
  container: 3000,
};
