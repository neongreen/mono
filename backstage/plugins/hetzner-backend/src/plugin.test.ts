import { startTestBackend } from '@backstage/backend-test-utils';
import { createServiceFactory } from '@backstage/backend-plugin-api';
import request from 'supertest';
import { hetznerPlugin } from './plugin';
import {
  HetznerServerSummary,
  HetznerSummary,
  hetznerServiceRef,
} from './services/HetznerService';

describe('plugin', () => {
  const serverSummary: HetznerServerSummary = {
    id: 99,
    name: 'worker-99',
    status: 'running',
    datacenter: 'nbg1-dc3',
    location: 'nbg1',
    ipv4Address: '198.51.100.10',
    ipv6Address: null,
    serverType: 'cx11',
    createdAt: new Date().toISOString(),
    locked: false,
    entityRef: 'resource:default/worker-99',
  };

  const summary: HetznerSummary = {
    project: {
      title: 'Hetzner Cloud',
      owner: 'user:default/neongreen',
      lifecycle: 'production',
      totals: {
        servers: 1,
        running: 1,
        datacenters: ['nbg1-dc3'],
      },
      generatedAt: new Date().toISOString(),
    },
    servers: [serverSummary],
  };

  it('exposes Hetzner REST endpoints', async () => {
    const hetznerService = {
      snapshot: jest.fn().mockResolvedValue(summary),
      listServers: jest.fn().mockResolvedValue(summary.servers),
      getServerById: jest.fn().mockResolvedValue(serverSummary),
    };

    const { server } = await startTestBackend({
      features: [
        hetznerPlugin,
        createServiceFactory({
          service: hetznerServiceRef,
          deps: {},
          factory: () =>
            hetznerService as unknown as typeof hetznerServiceRef.T,
        }),
      ],
    });

    await request(server)
      .get('/api/hetzner/summary')
      .expect(200, summary);

    await request(server)
      .get('/api/hetzner/servers')
      .expect(200, { servers: summary.servers });

    await request(server)
      .get('/api/hetzner/servers/99')
      .expect(200, serverSummary);
  });
});
