import {
  createPlugin,
  createRoutableExtension,
} from '@backstage/core-plugin-api';

import { rootRouteRef } from './routes';

export const hetznerPlugin = createPlugin({
  id: 'hetzner',
  routes: {
    root: rootRouteRef,
  },
});

export const HetznerPage = hetznerPlugin.provide(
  createRoutableExtension({
    name: 'HetznerPage',
    component: () =>
      import('./components/HetznerPage/HetznerPage').then(m => m.HetznerPage),
    mountPoint: rootRouteRef,
  }),
);
