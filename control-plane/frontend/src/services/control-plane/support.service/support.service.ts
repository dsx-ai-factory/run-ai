import { API } from "@/common/api.constant";
import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";

export const supportService = {
  contactUs,
};
const endpoint = `${API.v1}/support`;

async function contactUs(message: string): Promise<void> {
  return await controlPlaneService.post(`${endpoint}/contact`, { message });
}
