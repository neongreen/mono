import { createBackend } from '@backstage/backend-defaults';
import { mockServices } from '@backstage/backend-test-utils';
import { catalogServiceMock } from '@backstage/plugin-catalog-node/testUtils';

// TEMPLATE NOTE:
// This is the development setup for your plugin that wires up a
// minimal backend that can use both real and mocked plugins and services.
//
// Start up the backend by running `yarn start` in the package directory.
// Once it's up and running, try out the following requests:
//
// Fetch the Hetzner summary:
//   curl http://localhost:7007/api/hetzner/summary
//
// List servers:
//   curl http://localhost:7007/api/hetzner/servers
//
// Get a specific server (replace 1 with an id from the list response):
//   curl http://localhost:7007/api/hetzner/servers/1

const backend = createBackend();

// TEMPLATE NOTE:
// Mocking the auth and httpAuth service allows you to call your plugin API without
// having to authenticate.
//
// If you want to use real auth, you can install the following instead:
//   backend.add(import('@backstage/plugin-auth-backend'));
//   backend.add(import('@backstage/plugin-auth-backend-module-guest-provider'));
backend.add(mockServices.auth.factory());
backend.add(mockServices.httpAuth.factory());

// TEMPLATE NOTE:
// Rather than using a real catalog you can use a mock with a fixed set of entities.
backend.add(
  catalogServiceMock.factory({
    entities: [
      {
        apiVersion: 'backstage.io/v1alpha1',
        kind: 'Component',
        metadata: {
          name: 'sample',
          title: 'Sample Component',
        },
        spec: {
          type: 'service',
        },
      },
    ],
  }),
);

backend.add(import('../src'));

backend.start();
