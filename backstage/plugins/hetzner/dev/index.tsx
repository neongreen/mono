import { createDevApp } from '@backstage/dev-utils';
import { hetznerPlugin, HetznerPage } from '../src/plugin';

createDevApp()
  .registerPlugin(hetznerPlugin)
  .addPage({
    element: <HetznerPage />,
    title: 'Root Page',
    path: '/hetzner',
  })
  .render();
