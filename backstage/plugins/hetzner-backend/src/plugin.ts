import {
  coreServices,
  createBackendPlugin,
} from '@backstage/backend-plugin-api';
import { createRouter } from './router';
import { hetznerServiceRef } from './services/HetznerService';

/**
 * hetznerPlugin backend plugin
 *
 * @public
 */
export const hetznerPlugin = createBackendPlugin({
  pluginId: 'hetzner',
  register(env) {
    env.registerInit({
      deps: {
        httpAuth: coreServices.httpAuth,
        httpRouter: coreServices.httpRouter,
        hetznerService: hetznerServiceRef,
      },
      async init({ httpAuth, httpRouter, hetznerService }) {
        httpRouter.use(
          await createRouter({
            httpAuth,
            hetznerService,
          }),
        );
      },
    });
  },
});
