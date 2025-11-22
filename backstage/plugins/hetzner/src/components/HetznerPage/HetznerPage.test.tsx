import { HetznerPage } from './HetznerPage';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { screen } from '@testing-library/react';
import {
  registerMswTestHooks,
  renderInTestApp,
  TestApiProvider,
} from '@backstage/test-utils';
import { HetznerSummaryResponse } from '../../types';
import { discoveryApiRef } from '@backstage/core-plugin-api';
import { entityRouteRef } from '@backstage/plugin-catalog-react';

describe('HetznerPage', () => {
  const server = setupServer();
  // Enable sane handlers for network requests
  registerMswTestHooks(server);

  // setup mock response
  beforeEach(() => {
    const summary: HetznerSummaryResponse = {
      project: {
        title: 'Hetzner Cloud',
        owner: 'user:default/tester',
        lifecycle: 'production',
        totals: {
          servers: 1,
          running: 1,
          datacenters: ['hel1-dc1'],
        },
        generatedAt: new Date().toISOString(),
      },
      servers: [
        {
          id: 1,
          name: 'demo',
          status: 'running',
          datacenter: 'hel1-dc1',
          location: 'hel1',
          ipv4Address: '192.0.2.1',
          ipv6Address: null,
          serverType: 'cx21',
          createdAt: new Date().toISOString(),
          locked: false,
          entityRef: 'resource:default/demo',
        },
      ],
    };

    server.use(
      rest.get('http://localhost/api/hetzner/summary', (_, res, ctx) =>
        res(ctx.status(200), ctx.json(summary)),
      ),
    );
  });

  it('should render', async () => {
    await renderInTestApp(
      <TestApiProvider
        apis={[
          [
            discoveryApiRef,
            {
              async getBaseUrl(pluginId: string) {
                return `http://localhost/api/${pluginId}`;
              },
            },
          ],
        ]}
      >
        <HetznerPage />
      </TestApiProvider>,
      {
        mountedRoutes: {
          '/catalog/:namespace/:kind/:name': entityRouteRef,
        },
      },
    );
    expect(await screen.findByText('Hetzner Cloud')).toBeInTheDocument();
  });
});
