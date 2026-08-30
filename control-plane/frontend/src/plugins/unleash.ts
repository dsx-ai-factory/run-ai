import type { UnleashClient } from "@unleash/proxy-client-vue";
import { plugin as unleashPlugin } from "@unleash/proxy-client-vue";
import { unleashConfigService } from "@/services/infra/unleash.service/unleash-config.service";

import type { App } from "vue";

export { initUnleashClient };

function initUnleashClient(app: App): void {
  const unleashClient: UnleashClient | undefined = unleashConfigService.getClient();
  if (unleashClient) {
    app.use(unleashPlugin, { unleashClient });
  }
}
