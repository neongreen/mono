import {
  coreServices,
  createBackendModule,
  readSchedulerServiceTaskScheduleDefinitionFromConfig,
} from '@backstage/backend-plugin-api';
import { catalogProcessingExtensionPoint } from '@backstage/plugin-catalog-node/alpha';
import { HetznerEntityProvider } from './providers/HetznerEntityProvider';
import { hetznerServiceRef } from './services/HetznerService';

export const catalogModuleHetznerProvider = createBackendModule({
  pluginId: 'catalog',
  moduleId: 'hetzner-entity-provider',
  register(env) {
    env.registerInit({
      deps: {
        catalogProcessing: catalogProcessingExtensionPoint,
        config: coreServices.rootConfig,
        hetznerService: hetznerServiceRef,
        logger: coreServices.logger,
        scheduler: coreServices.scheduler,
      },
      async init({
        catalogProcessing,
        config,
        hetznerService,
        logger,
        scheduler,
      }) {
        const provider = new HetznerEntityProvider({
          logger: logger.child({ module: 'HetznerEntityProvider' }),
          service: hetznerService,
        });

        catalogProcessing.addEntityProvider(provider);

        const scheduleConfig = config
          .getOptionalConfig('backend.hetzner.entityProvider.schedule');
        const schedule = scheduleConfig
          ? readSchedulerServiceTaskScheduleDefinitionFromConfig(scheduleConfig)
          : {
              frequency: { minutes: 10 },
              timeout: { minutes: 2 },
            };

        await scheduler.scheduleTask({
          id: 'hetzner-cloud-refresh',
          ...schedule,
          fn: async () => {
            await provider.run();
          },
        });
      },
    });
  },
});
