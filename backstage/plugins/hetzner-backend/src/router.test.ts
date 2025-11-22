import {
  mockCredentials,
  mockErrorHandler,
  mockServices,
} from '@backstage/backend-test-utils';
import express from 'express';
import request from 'supertest';

import { createRouter } from './router';
import {
  HetznerServerSummary,
  HetznerSummary,
  hetznerServiceRef,
} from './services/HetznerService';

describe('createRouter', () => {
  let app: express.Express;
  let hetznerService: jest.Mocked<Pick<
    typeof hetznerServiceRef.T,
    'snapshot' | 'listServers' | 'getServerById'
  >>;

  const sampleServer: HetznerServerSummary = {
    id: 1,
    name: 'demo-server',
    status: 'running',
    datacenter: 'hel1-dc1',
    location: 'hel1',
    ipv4Address: '192.0.2.1',
    ipv6Address: null,
    serverType: 'cx21',
    createdAt: new Date().toISOString(),
    locked: false,
    entityRef: 'resource:default/demo-server',
  };

  const summary: HetznerSummary = {
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
    servers: [sampleServer],
  };

  beforeEach(async () => {
    hetznerService = {
      snapshot: jest.fn().mockResolvedValue(summary),
      listServers: jest.fn().mockResolvedValue(summary.servers),
      getServerById: jest.fn().mockResolvedValue(sampleServer),
    };
    const router = await createRouter({
      httpAuth: mockServices.httpAuth(),
      hetznerService: hetznerService as unknown as typeof hetznerServiceRef.T,
    });
    app = express();
    app.use(router);
    app.use(mockErrorHandler());
  });

  it('returns a project summary', async () => {
    const response = await request(app).get('/summary');

    expect(response.status).toBe(200);
    expect(response.body).toEqual(summary);
    expect(hetznerService.snapshot).toHaveBeenCalled();
  });

  it('lists servers', async () => {
    const response = await request(app).get('/servers');

    expect(response.status).toBe(200);
    expect(response.body).toEqual({ servers: summary.servers });
    expect(hetznerService.listServers).toHaveBeenCalled();
  });

  it('returns a single server', async () => {
    const response = await request(app).get('/servers/1');

    expect(response.status).toBe(200);
    expect(response.body).toEqual(sampleServer);
    expect(hetznerService.getServerById).toHaveBeenCalledWith(1);
  });

  it('validates the server id', async () => {
    const response = await request(app).get('/servers/not-a-number');
    expect(response.status).toBe(400);
  });

  it('requires authentication', async () => {
    const response = await request(app)
      .get('/summary')
      .set('Authorization', mockCredentials.none.header());

    expect(response.status).toBe(401);
  });
});
